package scaffold

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/jani123/dome/core/fault"
	"golang.org/x/sys/unix"
)

// Runner isolates dependency resolution from filesystem publication.
type Runner func(context.Context, string) error

// Tidy resolves the generated module. Command output is not included in errors:
// private module authentication and proxy URLs may contain credentials.
func Tidy(ctx context.Context, dir string) error {
	cmd := exec.CommandContext(ctx, "go", "mod", "tidy")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		return fault.Wrap(fault.Configuration, err)
	}
	return nil
}

// Publish creates a fresh directory atomically after successful dependency
// resolution. Existing destinations and symlinked parent paths are refused.
// Linux renameat2(RENAME_NOREPLACE) prevents overwriting a concurrent creator.
func Publish(ctx context.Context, destination string, files map[string][]byte, resolve Runner) (err error) {
	defer func() { err = fault.Wrap(fault.Configuration, err) }()
	destination, err = filepath.Abs(destination)
	if err != nil {
		return err
	}
	parent := filepath.Dir(destination)
	for path := parent; ; path = filepath.Dir(path) {
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fault.New(fault.Validation, "destination ancestors must be existing real directories")
		}
		if path == filepath.Dir(path) {
			break
		}
	}
	if _, err := os.Lstat(destination); !errors.Is(err, fs.ErrNotExist) {
		return fault.New(fault.Validation, "destination already exists or cannot be inspected")
	}
	stage, err := os.MkdirTemp(parent, ".dome-init-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if !filepath.IsLocal(name) {
			return fault.New(fault.Validation, "invalid template path")
		}
		path := filepath.Join(stage, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return err
		}
		if err := os.WriteFile(path, files[name], 0o600); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(filepath.Join(stage, "sandbox"), 0o700); err != nil {
		return err
	}
	if resolve != nil {
		if err := resolve(ctx, stage); err != nil {
			return err
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return unix.Renameat2(unix.AT_FDCWD, stage, unix.AT_FDCWD, destination, unix.RENAME_NOREPLACE)
}
