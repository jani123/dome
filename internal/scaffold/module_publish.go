package scaffold

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jani123/dome/core/fault"
	"golang.org/x/sys/unix"
)

// PrepareModule resolves dependencies and refreshes the pinned OpenSpec assets
// in private staging. Subprocess output is deliberately omitted from errors.
func PrepareModule(ctx context.Context, dir string) error {
	if err := Tidy(ctx, dir); err != nil {
		return err
	}
	for _, script := range []string{"scripts/install-openspec.sh", "scripts/setup-specs.sh"} {
		cmd := exec.CommandContext(ctx, "bash", script)
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			return fault.Wrap(fault.Configuration, err)
		}
	}
	return os.RemoveAll(filepath.Join(dir, ".build"))
}

// openRealDir retains the destination identity and rejects symlink traversal at
// every ancestor. Callers own the returned descriptor.
func openRealDir(path string) (int, error) {
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
		_ = unix.Close(fd)
		if e != nil {
			return -1, e
		}
		fd = next
	}
	return fd, nil
}

// PublishModule prepares a library privately, then publishes entries without
// replacing the existing empty working directory or overwriting concurrent files.
// Abrupt process death can leave partial output; reruns refuse nonempty output.
func PublishModule(ctx context.Context, destination string, files map[string]File, prepare Runner) (err error) {
	defer func() { err = fault.Wrap(fault.Configuration, err) }()
	fd, err := openRealDir(destination)
	if err != nil {
		return fault.Wrap(fault.Validation, err)
	}
	dir := os.NewFile(uintptr(fd), "module destination")
	defer dir.Close()
	empty := func() error {
		_, err := dir.Seek(0, 0)
		if err != nil {
			return err
		}
		names, err := dir.Readdirnames(1)
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}
		if len(names) != 0 {
			return fault.New(fault.Validation, "module initialization requires an existing empty directory, including no hidden entries")
		}
		return nil
	}
	if err := empty(); err != nil {
		return err
	}
	before, err := dir.Stat()
	if err != nil {
		return err
	}
	absolute, err := filepath.Abs(destination)
	if err != nil {
		return err
	}
	parentFD, err := openRealDir(filepath.Dir(absolute))
	if err != nil {
		return err
	}
	parent := os.NewFile(uintptr(parentFD), "module staging parent")
	defer parent.Close()
	stage, err := os.MkdirTemp("/proc/self/fd/"+strconv.Itoa(parentFD), ".dome-module-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	for name, file := range files {
		if !filepath.IsLocal(name) {
			return fault.New(fault.Validation, "invalid template path")
		}
		path := filepath.Join(stage, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, file.Data, file.Mode); err != nil {
			return err
		}
	}
	if prepare != nil {
		if err := prepare(ctx, stage); err != nil {
			return err
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	check, err := openRealDir(destination)
	if err != nil {
		return err
	}
	checkFile := os.NewFile(uintptr(check), "module destination check")
	info, err := checkFile.Stat()
	_ = checkFile.Close()
	if err != nil {
		return err
	}
	if !os.SameFile(before, info) {
		return fault.New(fault.Validation, "destination changed during preparation")
	}
	if err := empty(); err != nil {
		return err
	}
	stageDir, err := os.Open(stage)
	if err != nil {
		return err
	}
	defer stageDir.Close()
	// Reject staged symlinks before exposing any generated entry.
	if err := filepath.WalkDir(stage, func(_ string, e os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if e.Type()&os.ModeSymlink != 0 {
			return fault.New(fault.Validation, "staged symlinks are not supported")
		}
		return nil
	}); err != nil {
		return err
	}
	entries, err := os.ReadDir(stage)
	if err != nil {
		return err
	}
	published := []string{}
	// Never delete published paths during failure cleanup. Another writer may
	// replace an inode after an ownership check, and Linux has no conditional
	// unlink-by-inode. Retain partial output rather than risk another writer's data.
	defer func() {
		if err != nil && len(published) > 0 {
			err = fault.Wrap(fault.Configuration, err, "cleanup", "partial output retained; inspect destination before retry")
		}
	}()
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := unix.Renameat2(int(stageDir.Fd()), entry.Name(), fd, entry.Name(), unix.RENAME_NOREPLACE); err != nil {
			return err
		}
		published = append(published, entry.Name())
	}
	return nil
}
