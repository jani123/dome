package workspace_test

import (
	"testing"

	"github.com/jani123/dome/workspace"
	"github.com/stretchr/testify/require"
)

func TestStrictProfile(t *testing.T) {
	for _, body := range []string{
		"schema_version: 1\nsource: local\nconfig: config.yaml\nprivate_key: SECRET",
		"schema_version: [SECRET]", "schema_version: 1\n---\nsource: local", "",
	} {
		_, err := workspace.Decode([]byte(body))
		require.Error(t, err)
		require.NotContains(t, err.Error(), "SECRET")
	}
	for _, p := range []workspace.Profile{
		{SchemaVersion: 2, Source: "local", Config: "config.yaml"},
		{SchemaVersion: 1, Source: "local", Config: "config.yaml", GitURL: "https://example.com/x"},
		{SchemaVersion: 1, Source: "git", GitURL: "https://user:SECRET@example.com/x"},
		{SchemaVersion: 1, Source: "git", GitURL: "https://SECRET@example.com/x"},
		{SchemaVersion: 1, Source: "git", GitURL: "https://example.com/x", GitPath: "../config.yaml"},
		{SchemaVersion: 1, Source: "git", GitURL: "ssh://git@example.com/x"},
		{SchemaVersion: 1, Source: "git", GitURL: "http://example.com/x"},
	} {
		err := p.Defaults().Validate()
		require.Error(t, err)
		require.NotContains(t, err.Error(), "SECRET")
	}
}

func TestPathsAndDefaults(t *testing.T) {
	p := workspace.Profile{SchemaVersion: 1, Source: "git", GitURL: "ssh://git@example.com/x", SSHKey: "key", KnownHosts: "hosts", Secrets: "secrets.enc.yaml"}.Defaults()
	require.NoError(t, p.Validate())
	p = p.Resolve("/workspace")
	require.Equal(t, "/workspace/key", p.SSHKey)
	require.Equal(t, "/workspace/hosts", p.KnownHosts)
	require.Equal(t, "secrets.enc.yaml", p.Secrets)
	require.Equal(t, "config.yaml", p.GitPath)
	require.Equal(t, "HEAD", p.GitRef)
	require.Equal(t, "/workspace/.dome/state.db", p.State)
	p = workspace.Profile{SchemaVersion: 1, Source: "local", Config: "config.yaml", Secrets: "secret.yaml", State: "/state.db"}.Resolve("/workspace")
	require.NoError(t, p.Validate())
	require.Equal(t, "/workspace/secret.yaml", p.Secrets)
	require.Equal(t, "/state.db", p.State)
}
