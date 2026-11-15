//go:build integration

package workspace

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

func TestDirectoryReplacementBeforePublication(t *testing.T) {
	for _, ancestor := range []bool{false, true} {
		t.Run(map[bool]string{false: "profile-directory", true: "workspace-ancestor"}[ancestor], func(t *testing.T) {
			parent := t.TempDir()
			dir := filepath.Join(parent, "workspace")
			require.NoError(t, os.MkdirAll(filepath.Join(dir, ".dome"), 0o700))
			p := Profile{SchemaVersion: 1, Source: "git", GitURL: "https://example.com/x"}.Defaults()
			moved := filepath.Join(parent, "moved")
			err := (OSStore{}).save(context.Background(), dir, p, func() {
				original := filepath.Join(dir, ".dome")
				if ancestor {
					original = dir
				}
				require.NoError(t, os.Rename(original, moved))
				require.NoError(t, os.MkdirAll(filepath.Join(dir, ".dome"), 0o700))
				require.NoError(t, os.WriteFile(filepath.Join(dir, ".dome/agent.yaml"), []byte("concurrent writer"), 0o600))
			})
			require.ErrorIs(t, err, unix.ESTALE)
			data, readErr := os.ReadFile(filepath.Join(dir, ".dome/agent.yaml"))
			require.NoError(t, readErr)
			require.Equal(t, "concurrent writer", string(data))
			if ancestor {
				moved = filepath.Join(moved, ".dome")
			}
			entries, readErr := os.ReadDir(moved)
			require.NoError(t, readErr)
			require.Empty(t, entries)
		})
	}
}

func TestAnonymousStageHasNoMutableNameOrCleanup(t *testing.T) {
	dir := t.TempDir()
	p := Profile{SchemaVersion: 1, Source: "git", GitURL: "https://example.com/x"}.Defaults()
	err := (OSStore{}).save(context.Background(), dir, p, func() {
		entries, err := os.ReadDir(filepath.Join(dir, ".dome"))
		require.NoError(t, err)
		require.Empty(t, entries)
		// An unrelated concurrent file must not be removed by staging cleanup.
		require.NoError(t, os.WriteFile(filepath.Join(dir, ".dome/unrelated"), []byte("keep"), 0o600))
	})
	require.NoError(t, err)
	data, err := os.ReadFile(filepath.Join(dir, ".dome/unrelated"))
	require.NoError(t, err)
	require.Equal(t, "keep", string(data))
	_, found, err := (OSStore{}).Load(context.Background(), dir, "")
	require.NoError(t, err)
	require.True(t, found)
}
