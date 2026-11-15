package workspace

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jani123/dome/core/fault"
	"golang.org/x/sys/unix"
	"gopkg.in/yaml.v3"
)

// Store isolates host I/O for unit tests. Load returns false only for absence.
type Store interface {
	Load(context.Context, string, string) (Profile, bool, error)
	Save(context.Context, string, Profile) error
	CheckFile(context.Context, string) error
}

// OSStore publishes profiles using Linux descriptor-relative, no-follow
// operations. The destination filesystem must support O_TMPFILE and hard links;
// /proc/self/fd must be available. Unsupported capabilities return errors.
type OSStore struct{}

type concealedCause struct {
	cause   error
	message string
}

func (e concealedCause) Error() string { return e.message }
func (e concealedCause) Unwrap() error { return e.cause }
func concealed(err error, message string) error {
	return fault.Wrap(fault.Configuration, concealedCause{cause: err, message: message})
}

func openDir(path string, create bool) (int, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return -1, err
	}
	fd, err := unix.Open("/", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return -1, err
	}
	for _, part := range strings.Split(strings.TrimPrefix(absolute, "/"), "/") {
		if part == "" {
			continue
		}
		next, e := unix.Openat(fd, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
		if errors.Is(e, unix.ENOENT) && create {
			e = unix.Mkdirat(fd, part, 0o700)
			if e == nil || errors.Is(e, unix.EEXIST) {
				next, e = unix.Openat(fd, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
			}
		}
		_ = unix.Close(fd)
		if e != nil {
			return -1, e
		}
		fd = next
	}
	return fd, nil
}

// Load reads either the default profile or an explicit file, rejecting symlinks.
func (OSStore) Load(ctx context.Context, dir, path string) (Profile, bool, error) {
	if err := ctx.Err(); err != nil {
		return Profile{}, false, err
	}
	if path == "" {
		path = filepath.Join(dir, ".dome/agent.yaml")
	} else if !filepath.IsAbs(path) {
		path = filepath.Join(dir, path)
	}
	fd, err := openDir(filepath.Dir(path), false)
	if errors.Is(err, unix.ENOENT) {
		return Profile{}, false, nil
	}
	if err != nil {
		return Profile{}, false, concealed(err, "cannot open workspace profile directory")
	}
	defer unix.Close(fd)
	f, err := unix.Openat(fd, filepath.Base(path), unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
	if errors.Is(err, unix.ENOENT) {
		return Profile{}, false, nil
	}
	if err != nil {
		return Profile{}, false, concealed(err, "cannot open workspace profile")
	}
	file := os.NewFile(uintptr(f), "workspace profile")
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return Profile{}, false, concealed(err, "cannot inspect workspace profile")
	}
	if !info.Mode().IsRegular() {
		return Profile{}, false, invalid("workspace profile must be a regular file")
	}
	data, err := io.ReadAll(io.LimitReader(file, 64<<10+1))
	if err != nil {
		return Profile{}, false, concealed(err, "cannot read workspace profile")
	}
	p, err := Decode(data)
	if err != nil {
		return Profile{}, false, err
	}
	if err = ctx.Err(); err != nil {
		return Profile{}, false, err
	}
	return p, true, nil
}

// CheckFile verifies a readable regular reference without retaining its contents.
func (OSStore) CheckFile(ctx context.Context, path string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	descriptor, err := unix.Open(path, unix.O_RDONLY|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
	if err != nil {
		return concealed(err, "referenced configuration or credential file is not readable")
	}
	f := os.NewFile(uintptr(descriptor), "workspace reference")
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return concealed(err, "cannot inspect workspace reference")
	}
	if !info.Mode().IsRegular() {
		return invalid("referenced configuration or credential must be a regular file")
	}
	return nil
}

func sameDirectory(fd int, path string) error {
	current, err := openDir(path, false)
	if err != nil {
		return concealed(err, "workspace directory changed during initialization")
	}
	defer unix.Close(current)
	var held, visible unix.Stat_t
	if err = unix.Fstat(fd, &held); err != nil {
		return concealed(err, "cannot inspect held workspace directory")
	}
	if err = unix.Fstat(current, &visible); err != nil {
		return concealed(err, "cannot inspect visible workspace directory")
	}
	if held.Dev != visible.Dev || held.Ino != visible.Ino {
		return concealed(unix.ESTALE, "workspace directory changed during initialization")
	}
	return nil
}

// Save atomically publishes a complete profile. An equivalent existing profile is
// retained; conflicting or concurrently created output is never overwritten.
func (s OSStore) Save(ctx context.Context, dir string, p Profile) error {
	return s.save(ctx, dir, p, nil)
}

// staged is an internal integration-test seam for deterministic directory races.
func (s OSStore) save(ctx context.Context, dir string, p Profile, staged func()) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	p = p.Defaults()
	if err := p.Validate(); err != nil {
		return err
	}
	data, err := yaml.Marshal(p)
	if err != nil {
		return concealed(err, "cannot encode workspace profile")
	}
	if len(data) > 64<<10 {
		return invalid("workspace profile exceeds 64 KiB")
	}
	existing, found, err := s.Load(ctx, dir, "")
	if err != nil {
		return err
	}
	if found {
		if existing.Defaults().Resolve(dir) == p.Resolve(dir) {
			return nil
		}
		return invalid("workspace already initialized with different settings; edit its profile deliberately")
	}
	path := filepath.Join(dir, ".dome")
	fd, err := openDir(path, true)
	if err != nil {
		return concealed(err, "cannot create workspace profile directory")
	}
	defer unix.Close(fd)
	// Anonymous staging has no mutable source pathname and no pathname cleanup.
	// Linking /proc/self/fd pins the written inode without requiring CAP_DAC_READ_SEARCH.
	f, err := unix.Openat(fd, ".", unix.O_WRONLY|unix.O_TMPFILE|unix.O_CLOEXEC, 0o600)
	if err != nil {
		return concealed(err, "cannot create anonymous workspace staging file; filesystem must support O_TMPFILE")
	}
	file := os.NewFile(uintptr(f), "workspace staging")
	defer file.Close()
	if _, err = file.Write(data); err != nil {
		return concealed(err, "cannot write workspace profile")
	}
	if err = file.Sync(); err != nil {
		return concealed(err, "cannot sync workspace profile")
	}
	if staged != nil {
		staged()
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	if err = sameDirectory(fd, path); err != nil {
		return err
	}
	if err = unix.Linkat(unix.AT_FDCWD, "/proc/self/fd/"+strconv.Itoa(f), fd, "agent.yaml", unix.AT_SYMLINK_FOLLOW); err != nil {
		return concealed(err, "workspace profile publication refused; check destination and filesystem hard-link/proc support")
	}
	// Never unlink visible output after publication: another writer may replace it.
	if err = sameDirectory(fd, path); err != nil {
		return err
	}
	if err = unix.Fsync(fd); err != nil {
		return concealed(err, "workspace profile published but directory sync failed")
	}
	return nil
}
