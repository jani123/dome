//go:build integration

package knowledge_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jani123/dome/internal/knowledge"
	"github.com/stretchr/testify/require"
)

func TestRootAndSymlinkBoundaries(t *testing.T) {
	dir := t.TempDir()
	for name, body := range fixture() {
		target := filepath.Join(dir, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(target), 0o700))
		require.NoError(t, os.WriteFile(target, []byte(body), 0o600))
	}
	root, err := os.OpenRoot(dir)
	require.NoError(t, err)
	defer root.Close()
	issues, err := knowledge.Validate(root.FS(), knowledge.Options{Bundle: "knowledge", Policy: true})
	require.NoError(t, err)
	require.Empty(t, issues)
	outside := filepath.Join(t.TempDir(), "private.md")
	require.NoError(t, os.WriteFile(outside, []byte("private contents"), 0o600))
	require.NoError(t, os.Symlink(outside, filepath.Join(dir, "knowledge", "escape.md")))
	issues, err = knowledge.Validate(root.FS(), knowledge.Options{Bundle: "knowledge", Policy: true})
	require.NoError(t, err)
	require.NotEmpty(t, issues)
	for _, issue := range issues {
		require.NotContains(t, issue.String(), "private contents")
	}
}
