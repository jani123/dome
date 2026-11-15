package scaffold_test

import (
	"encoding/json"
	"go/parser"
	"go/token"
	"testing"

	"github.com/jani123/dome/internal/scaffold"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestRender(t *testing.T) {
	for _, version := range []string{"v0.1.0-rc.1", "v0.1.0", "v0.2.0", "v0.0.0-20261112182025-28339340120a", "v0.1.1-0.20261112182025-28339340120a"} {
		t.Run(version, func(t *testing.T) {
			files, err := scaffold.Render("example.com/my-machine", version)
			require.NoError(t, err)
			require.Contains(t, string(files["go.mod"]), "github.com/jani123/dome "+version)
			_, err = parser.ParseFile(token.NewFileSet(), "main.go", files["cmd/agent/main.go"], parser.AllErrors)
			require.NoError(t, err)
			var manifest map[string]any
			require.NoError(t, json.Unmarshal(files["dome.json"], &manifest))
			for _, field := range []string{"dome_version", "skill_version", "template_version"} {
				require.Equal(t, version, manifest[field])
			}
			require.NoError(t, scaffold.CheckVersion(version, files["dome.json"]))
			require.Error(t, scaffold.CheckVersion("v0.3.0", files["dome.json"]))
			skill := string(files[".agents/skills/dome/SKILL.md"])
			require.Contains(t, skill, "go list -m -f '{{.Version}}'")
			require.Contains(t, skill, version)
			var front map[string]any
			require.NoError(t, yaml.Unmarshal([]byte(skill[4:findFrontmatterEnd(skill)]), &front))
			require.Equal(t, "dome", front["name"])
			require.NotEmpty(t, front["description"])
		})
	}
}

func findFrontmatterEnd(s string) int {
	for i := 4; i+4 <= len(s); i++ {
		if s[i:i+4] == "---\n" {
			return i
		}
	}
	return 4
}

func TestInvalidInput(t *testing.T) {
	for _, module := range []string{"", "bad path", "../escape", "github.com/jani123/dome", "github.com/jani123/dome/consumer"} {
		_, err := scaffold.Render(module, "v0.1.0")
		require.Error(t, err)
	}
	for _, version := range []string{"dev", "main", "latest", "v1.0.0", "v0.1.0+build", "v0.1", "v0.0.0-20261112182025-28339340120a+build"} {
		_, err := scaffold.Render("example.com/consumer", version)
		require.Error(t, err)
	}
	require.Error(t, scaffold.CheckVersion("v0.1.0", []byte("invalid")))
}

func TestPatchedGoBaseline(t *testing.T) {
	files, err := scaffold.Render("example.com/consumer", "v0.1.0")
	require.NoError(t, err)
	require.Contains(t, string(files["go.mod"]), "go 1.26.11")
}

func TestConsumerDevcontainer(t *testing.T) {
	files, err := scaffold.Render("example.com/consumer", "v0.1.0")
	require.NoError(t, err)
	var cfg map[string]any
	require.NoError(t, json.Unmarshal(files[".devcontainer/devcontainer.json"], &cfg))
	require.Equal(t, "dev", cfg["remoteUser"])
	require.Equal(t, true, cfg["updateRemoteUserUID"])
	require.Equal(t, "/workspace", cfg["workspaceFolder"])
	require.Equal(t, "source=${localWorkspaceFolder},target=/workspace,type=bind", cfg["workspaceMount"])
	require.NotContains(t, cfg, "postCreateCommand")
	require.NotContains(t, cfg, "postStartCommand")
	require.NotContains(t, string(files[".devcontainer/devcontainer.json"]), "docker.sock")
	dockerfile := string(files[".devcontainer/Dockerfile"])
	require.Contains(t, dockerfile, "FROM golang:1.26.11-bookworm")
	require.Contains(t, dockerfile, "useradd")
	require.Contains(t, dockerfile, "USER dev")
	require.NotContains(t, files, "_devcontainer/devcontainer.json")
}
