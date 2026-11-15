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
	for _, version := range []string{"v0.1.0-rc.1", "v0.0.0-20261112182025-28339340120a"} {
		t.Run(version, func(t *testing.T) { testGeneratedConsumer(t, version) })
	}
}

func testGeneratedConsumer(t *testing.T, version string) {
	dest := filepath.Join(t.TempDir(), "consumer")
	files, err := scaffold.Render("example.com/consumer", version)
	require.NoError(t, err)
	// Exercise the local implementation explicitly; this does not prove public resolution.
	require.NoError(t, scaffold.CheckVersion(version, files["dome.json"]))
	files["go.mod"] = append(files["go.mod"], []byte("\nreplace github.com/jani123/dome => "+strconv.Quote(testutil.Root())+"\n")...)
	require.NoError(t, scaffold.Publish(context.Background(), dest, files, scaffold.Tidy))
	license, err := os.ReadFile(filepath.Join(testutil.Root(), "LICENSE"))
	require.NoError(t, err)
	require.Equal(t, string(license), string(files["LICENSE"]))
	require.FileExists(t, filepath.Join(dest, "sandbox/.gitkeep"))
	require.FileExists(t, filepath.Join(dest, ".devcontainer/devcontainer.json"))
	require.FileExists(t, filepath.Join(dest, ".devcontainer/Dockerfile"))
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
	require.Contains(t, run(binary, "init", "--config", "config.yaml", "--non-interactive"), "Workspace initialized")
	require.NoFileExists(t, filepath.Join(dest, ".dome/state.db"))
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

// TestGeneratedExternalModuleConsumer composes a separately generated public module.
// Local replacements exercise current sources without claiming public availability.
func TestGeneratedExternalModuleConsumer(t *testing.T) {
	ctx := context.Background()
	root := testutil.Root()
	moduleDir := t.TempDir()
	moduleFiles, err := scaffold.RenderModule("example.com/widget", "widget", "v0.1.0-rc.1")
	require.NoError(t, err)
	prepareModule := func(ctx context.Context, stage string) error {
		cmd := exec.CommandContext(ctx, "go", "mod", "edit", "-replace=github.com/jani123/dome="+root)
		cmd.Dir = stage
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
		return scaffold.Tidy(ctx, stage)
	}
	require.NoError(t, scaffold.PublishModule(ctx, moduleDir, moduleFiles, prepareModule))
	files, err := scaffold.RenderWithModules("example.com/consumer", "v0.1.0-rc.1", []string{"example.com/widget@v0.1.0"})
	require.NoError(t, err)
	files["go.mod"] = append(files["go.mod"], []byte("\nreplace github.com/jani123/dome => "+strconv.Quote(root)+"\nreplace example.com/widget => "+strconv.Quote(moduleDir)+"\n")...)
	dest := filepath.Join(t.TempDir(), "consumer")
	require.NoError(t, scaffold.Publish(ctx, dest, files, scaffold.WithPinnedDependencies(scaffold.Tidy, files["go.mod"])))
	run := func(name string, args ...string) string {
		t.Helper()
		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Dir = dest
		cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
		return string(out)
	}
	run("go", "build", "-o", "agent", "./cmd/agent")
	run("go", "test", "./...")
	binary := filepath.Join(dest, "agent")
	modules := run(binary, "modules")
	require.Contains(t, modules, "widget")
	require.Contains(t, modules, "file")
	run(binary, "init", "--config", "config.yaml", "--non-interactive")
	require.Contains(t, run(binary, "plan", "--json"), `"create"`)
	require.NoFileExists(t, filepath.Join(dest, "sandbox/hello.txt"))
	require.Contains(t, run(binary, "apply", "--json"), `"create"`)
	require.Contains(t, run(binary, "apply", "--json"), `"no-op"`)
}
