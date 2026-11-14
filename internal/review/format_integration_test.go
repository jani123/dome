//go:build integration

package review

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadOnlyFormatCheck(t *testing.T) {
	script, err := os.ReadFile("../../scripts/check-format.sh")
	require.NoError(t, err)
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "check.sh"), script, 0o755))
	cmd := exec.Command("git", "init", "-q")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())
	file := filepath.Join(dir, "outside normal dirs", "new file.go")
	require.NoError(t, os.MkdirAll(filepath.Dir(file), 0o755))
	clean := []byte("package fixture\n\nfunc Value() int { return 1 }\n")
	require.NoError(t, os.WriteFile(file, clean, 0o644))
	run := func() error { cmd := exec.Command("bash", "check.sh"); cmd.Dir = dir; return cmd.Run() }
	require.NoError(t, run())
	bad := []byte("package fixture\nfunc Value()int{return 1}\n")
	require.NoError(t, os.WriteFile(file, bad, 0o644))
	require.Error(t, run())
	got, err := os.ReadFile(file)
	require.NoError(t, err)
	require.Equal(t, bad, got)
	require.NoError(t, os.WriteFile(file, clean, 0o644))
	cmd = exec.Command("bash", "check.sh")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "PATH=/usr/bin:/bin")
	require.Error(t, cmd.Run(), "missing formatter must not look like a successful empty result")
}
