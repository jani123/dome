//go:build integration

package scaffold

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPinnedResolutionPreventsPublication(t *testing.T) {
	files, err := RenderWithModules("example.com/consumer", "v0.1.0", []string{"example.com/widget@v1.0.0"})
	require.NoError(t, err)
	failure := errors.New("resolution failed")
	for _, tc := range []struct {
		name   string
		runner Runner
	}{
		{"changed pin", func(_ context.Context, dir string) error {
			return os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/consumer\nrequire github.com/jani123/dome v0.2.0\nrequire example.com/widget v1.0.0\n"), 0o600)
		}},
		{"resolution error", func(context.Context, string) error { return failure }},
		{"missing resolver", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			parent := t.TempDir()
			dest := filepath.Join(parent, "consumer")
			err := Publish(context.Background(), dest, files, WithPinnedDependencies(tc.runner, files["go.mod"]))
			require.Error(t, err)
			if tc.name == "resolution error" {
				require.ErrorIs(t, err, failure)
			}
			require.NoDirExists(t, dest)
			entries, err := os.ReadDir(parent)
			require.NoError(t, err)
			require.Empty(t, entries)
		})
	}
}
