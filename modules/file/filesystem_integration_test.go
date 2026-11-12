//go:build integration

package file_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/jani123/dome/modules/file"
	"github.com/stretchr/testify/require"
)

func TestOSAtomicPermissionsAndSymlinks(t *testing.T) {
	p := filepath.Join(t.TempDir(), "file")
	disk := file.OSFilesystem{}
	require.NoError(t, disk.WriteAtomic(p, []byte("a"), 0o640))
	require.NoError(t, disk.WriteAtomic(p, []byte("b"), 0o600))
	b, mode, err := disk.Read(p)
	require.NoError(t, err)
	require.Equal(t, []byte("b"), b)
	require.Equal(t, fs.FileMode(0o600), mode)
	link := p + "-link"
	require.NoError(t, os.Symlink(p, link))
	require.Error(t, disk.WriteAtomic(link, []byte("bad"), 0o600))
	require.Error(t, disk.Remove(link))
	require.NoError(t, disk.Remove(p))
	require.NoError(t, disk.Remove(p))
}
