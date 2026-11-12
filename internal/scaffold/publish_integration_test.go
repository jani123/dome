//go:build integration

package scaffold_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/jani123/dome/internal/scaffold"
	"github.com/stretchr/testify/require"
)

func TestPublishBoundaries(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	dest := filepath.Join(root, "consumer")
	files, err := scaffold.Render("example.com/consumer", "v0.1.0")
	require.NoError(t, err)
	failure := errors.New("resolution failed")
	require.ErrorIs(t, scaffold.Publish(ctx, dest, files, func(context.Context, string) error { return failure }), failure)
	entries, err := os.ReadDir(root)
	require.NoError(t, err)
	require.Empty(t, entries)
	require.NoError(t, scaffold.Publish(ctx, dest, files, nil))
	require.DirExists(t, filepath.Join(dest, "sandbox"))
	require.FileExists(t, filepath.Join(dest, ".agents/skills/dome/SKILL.md"))
	require.Error(t, scaffold.Publish(ctx, dest, files, nil))
	require.NoError(t, os.Symlink(root, filepath.Join(root, "link")))
	require.Error(t, scaffold.Publish(ctx, filepath.Join(root, "link", "other"), files, nil))
	require.Error(t, scaffold.Publish(ctx, filepath.Join(root, "escape"), map[string][]byte{"../outside": {}}, nil))
	require.NoFileExists(t, filepath.Join(root, "outside"))
	// Cancellation after dependency resolution must not publish partial output.
	cancelled, cancel := context.WithCancel(ctx)
	cancelledDest := filepath.Join(root, "cancelled")
	require.ErrorIs(t, scaffold.Publish(cancelled, cancelledDest, files, func(context.Context, string) error {
		cancel()
		return nil
	}), context.Canceled)
	require.NoDirExists(t, cancelledDest)
	// A concurrent creator must win without its files being overwritten.
	other := filepath.Join(root, "other")
	require.Error(t, scaffold.Publish(ctx, other, files, func(context.Context, string) error {
		require.NoError(t, os.Mkdir(other, 0o700))
		return nil
	}))
	entries, err = os.ReadDir(other)
	require.NoError(t, err)
	require.Empty(t, entries)
}
