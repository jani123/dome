package scaffold

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRenderModule(t *testing.T) {
	for _, name := range []string{"main", "init", "for", "Bad", "a-b", ""} {
		_, err := RenderModule("example.com/module", name, "v0.1.0")
		require.Error(t, err, name)
	}
	for _, path := range []string{"bad path", "github.com/jani123/dome/test", "example.com/module/v2"} {
		_, err := RenderModule(path, "demo", "v0.1.0")
		require.Error(t, err, path)
	}
	_, err := RenderModule("example.com/module", "demo", "dev")
	require.Error(t, err)
	files, err := RenderModule("example.com/module", "demo", "v0.1.0")
	require.NoError(t, err)
	for _, path := range []string{"resource.go", "Taskfile.yml", ".devcontainer/compose.yaml", ".devcontainer/devcontainer-lock.json", "devbox.lock", "docs/hugo.toml", "tests/e2e/e2e_test.go", "tests/integration/integration_test.go", ".agents/skills/dome-module/SKILL.md", ".agents/skills/openspec-verify-change/SKILL.md"} {
		require.Contains(t, files, path)
	}
	require.Contains(t, string(files[".devcontainer/proxy/squid.conf"].Data), "visible_hostname demo-egress")
	require.NotContains(t, files, "Makefile")
	require.NotContains(t, files, "sandbox/.gitkeep")
	require.EqualValues(t, 0o755, files["scripts/release.sh"].Mode)
	require.Contains(t, string(files["resource.go"].Data), "package demo")
	require.NotContains(t, string(files["internal/knowledge/validate.go"].Data), "github.com/jani123/dome/internal/")
}
