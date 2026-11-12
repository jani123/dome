//go:build integration

package integration_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/jani123/dome/internal/scaffold"
	"github.com/jani123/dome/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestGeneratedConsumer(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "consumer")
	files, err := scaffold.Render("example.com/consumer", "v0.1.0-rc.1")
	require.NoError(t, err)
	// Before the first tag exists, explicitly test against the local checkout.
	files["go.mod"] = append(files["go.mod"], []byte("\nreplace github.com/jani123/dome => "+strconv.Quote(testutil.Root())+"\n")...)
	require.NoError(t, scaffold.Publish(context.Background(), dest, files, scaffold.Tidy))
	license, err := os.ReadFile(filepath.Join(testutil.Root(), "LICENSE"))
	require.NoError(t, err)
	require.Equal(t, string(license), string(files["LICENSE"]))
	require.FileExists(t, filepath.Join(dest, "sandbox/.gitkeep"))
	run := func(name string, args ...string) string {
		t.Helper()
		cmd := exec.Command(name, args...)
		cmd.Dir = dest
		cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
		return string(out)
	}
	run("go", "build", "-o", "agent", "./cmd/agent")
	run("go", "test", "./...")
	binary := filepath.Join(dest, "agent")
	require.Contains(t, run(binary, "plan", "--json"), `"create"`)
	require.NoFileExists(t, filepath.Join(dest, "sandbox/hello.txt"))
	require.Contains(t, run(binary, "apply", "--json"), `"create"`)
	body, err := os.ReadFile(filepath.Join(dest, "sandbox/hello.txt"))
	require.NoError(t, err)
	require.Equal(t, "Hello from Dome\n", string(body))
	info, err := os.Stat(filepath.Join(dest, "sandbox/hello.txt"))
	require.NoError(t, err)
	require.EqualValues(t, 0o640, info.Mode().Perm())
	require.Contains(t, run(binary, "apply", "--json"), `"no-op"`)
	require.NoError(t, os.WriteFile(filepath.Join(dest, "config.yaml"), []byte("managed: true\ncontent: updated\n"), 0o600))
	require.Contains(t, run(binary, "apply", "--json"), `"update"`)
	body, err = os.ReadFile(filepath.Join(dest, "sandbox/hello.txt"))
	require.NoError(t, err)
	require.Equal(t, "updated", string(body))
	require.NoError(t, os.WriteFile(filepath.Join(dest, "sandbox/unrelated.txt"), []byte("keep"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dest, "config.yaml"), []byte("managed: false\nunknown: rejected\n"), 0o600))
	cmd := exec.Command(binary, "apply", "--json")
	cmd.Dir = dest
	require.Error(t, cmd.Run())
	require.FileExists(t, filepath.Join(dest, "sandbox/hello.txt"))
	require.NoError(t, os.WriteFile(filepath.Join(dest, "config.yaml"), []byte("managed: false\n"), 0o600))
	require.Contains(t, run(binary, "apply", "--json"), `"delete"`)
	require.NoFileExists(t, filepath.Join(dest, "sandbox/hello.txt"))
	require.FileExists(t, filepath.Join(dest, "sandbox/unrelated.txt"))
}
