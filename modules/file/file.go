// Package file manages regular files using atomic replacement.
package file

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/jani123/dome/core/fault"
	"github.com/jani123/dome/core/resource"
)

// Filesystem is the complete OS boundary for this module.
type (
	Filesystem interface {
		Read(string) ([]byte, fs.FileMode, error)
		WriteAtomic(string, []byte, fs.FileMode) error
		Remove(string) error
	}
	OSFilesystem struct{}
)

func (OSFilesystem) Read(path string) ([]byte, fs.FileMode, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, 0, err
	}
	if !info.Mode().IsRegular() {
		return nil, 0, fmt.Errorf("managed path is not a regular file")
	}
	b, err := os.ReadFile(path)
	return b, info.Mode().Perm(), err
}

func (OSFilesystem) WriteAtomic(path string, b []byte, mode fs.FileMode) error {
	if info, err := os.Lstat(path); err == nil && !info.Mode().IsRegular() {
		return fmt.Errorf("managed path is not a regular file")
	} else if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".dome-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	defer f.Close()
	if err = f.Chmod(mode); err != nil {
		return err
	}
	if _, err = f.Write(b); err != nil {
		return err
	}
	if err = f.Sync(); err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Rename(f.Name(), path); err != nil {
		return err
	}
	dir, err := os.Open(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}

func (OSFilesystem) Remove(path string) error {
	info, err := os.Lstat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("refusing to remove a non-regular file")
	}
	return os.Remove(path)
}

type File struct {
	Path    string
	Content []byte
	Mode    fs.FileMode
	Absent  bool
	FS      Filesystem
}

func (f *File) ID() resource.ID { return resource.ID{Module: "file", Name: f.Path} }
func (f *File) Validate() (err error) {
	defer func() { err = fault.Wrap(fault.Validation, err, "module", "file", "resource_id", f.ID().String()) }()
	if !filepath.IsAbs(f.Path) || filepath.Clean(f.Path) != f.Path || f.Path == "/" {
		return fmt.Errorf("file path must be a clean absolute file path")
	}
	if f.Mode & ^fs.FileMode(0o777) != 0 {
		return fmt.Errorf("file permissions must be between 0000 and 0777")
	}
	return nil
}

func (f *File) filesystem() Filesystem {
	if f.FS != nil {
		return f.FS
	}
	return OSFilesystem{}
}

func snapshot(b []byte, m fs.FileMode) resource.Snapshot {
	return resource.Snapshot{Exists: true, Value: fmt.Sprintf("sha256:%x;mode:%04o", sha256.Sum256(b), m.Perm())}
}

func (f *File) Desired() resource.Snapshot {
	if f.Absent {
		return resource.Snapshot{}
	}
	return snapshot(f.Content, f.Mode)
}

func (f *File) Observe(ctx context.Context) (_ resource.Snapshot, err error) {
	defer func() { err = fault.Wrap(fault.Observation, err, "module", "file", "resource_id", f.ID().String()) }()
	if err = f.Validate(); err != nil {
		return resource.Snapshot{}, err
	}
	if err := ctx.Err(); err != nil {
		return resource.Snapshot{}, err
	}
	b, m, err := f.filesystem().Read(f.Path)
	if errors.Is(err, fs.ErrNotExist) {
		return resource.Snapshot{}, nil
	}
	if err != nil {
		return resource.Snapshot{}, err
	}
	return snapshot(b, m), nil
}

func (f *File) Apply(ctx context.Context, action resource.Action) (err error) {
	defer func() { err = fault.Wrap(fault.Apply, err, "module", "file", "resource_id", f.ID().String()) }()
	if err = f.Validate(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if action == resource.Noop {
		return nil
	}
	if action == resource.Delete {
		return f.filesystem().Remove(f.Path)
	}
	if action != resource.Create && action != resource.Update {
		return fault.New(fault.Validation, "unsupported file action")
	}
	return f.filesystem().WriteAtomic(f.Path, f.Content, f.Mode)
}

type Module struct{ FS Filesystem }

func (Module) Name() string { return "file" }
func (m Module) Restore(id resource.ID) (resource.Resource, error) {
	if id.Module != "file" {
		return nil, fmt.Errorf("wrong module")
	}
	f := &File{Path: id.Name, Absent: true, FS: m.FS}
	return f, f.Validate()
}
