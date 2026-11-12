package file_test

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"testing"

	"github.com/jani123/dome/core/resource"
	"github.com/jani123/dome/modules/file"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type filesystem struct{ mock.Mock }

func (f *filesystem) Read(p string) ([]byte, fs.FileMode, error) {
	a := f.Called(p)
	return a.Get(0).([]byte), a.Get(1).(fs.FileMode), a.Error(2)
}

func (f *filesystem) WriteAtomic(p string, b []byte, m fs.FileMode) error {
	return f.Called(p, b, m).Error(0)
}
func (f *filesystem) Remove(p string) error { return f.Called(p).Error(0) }
func TestFileLifecycleWithMock(t *testing.T) {
	ctx := context.Background()
	disk := &filesystem{}
	f := &file.File{Path: "/managed", Content: []byte("secret"), Mode: 0o640, FS: disk}
	require.NoError(t, f.Validate())
	disk.On("Read", "/managed").Return([]byte(nil), fs.FileMode(0), fs.ErrNotExist).Once()
	actual, err := f.Observe(ctx)
	require.NoError(t, err)
	require.False(t, actual.Exists)
	require.NotContains(t, f.Desired().Value, "secret")
	disk.On("WriteAtomic", "/managed", []byte("secret"), fs.FileMode(0o640)).Return(nil).Once()
	require.NoError(t, f.Apply(ctx, resource.Create))
	disk.On("Read", "/managed").Return([]byte("secret"), fs.FileMode(0o640), nil).Once()
	actual, err = f.Observe(ctx)
	require.NoError(t, err)
	require.Equal(t, f.Desired(), actual)
	disk.On("Remove", "/managed").Return(nil).Once()
	require.NoError(t, f.Apply(ctx, resource.Delete))
	disk.AssertExpectations(t)
}

func TestFileValidationAndErrors(t *testing.T) {
	for _, p := range []string{"relative", "/a/../b", "/"} {
		require.Error(t, (&file.File{Path: p}).Validate())
	}
	require.Error(t, (&file.File{Path: "/a", Mode: os.ModeSetuid}).Validate())
	disk := &filesystem{}
	f := &file.File{Path: "/a", FS: disk}
	sentinel := errors.New("disk")
	disk.On("Read", "/a").Return([]byte(nil), fs.FileMode(0), sentinel)
	_, err := f.Observe(context.Background())
	require.ErrorIs(t, err, sentinel)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.ErrorIs(t, f.Apply(ctx, resource.Create), context.Canceled)
}
