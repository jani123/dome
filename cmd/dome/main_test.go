package main

import (
	"bytes"
	"context"
	"testing"

	"github.com/jani123/dome/internal/scaffold"
	"github.com/stretchr/testify/require"
)

func TestInitValidationAndPublication(t *testing.T) {
	for _, args := range [][]string{{"init"}, {"init", "dest"}, {"init", "dest", "--module", "bad path"}} {
		root := command("v0.1.0", func(context.Context, string, map[string][]byte, scaffold.Runner) error {
			t.Fatal("invalid request published")
			return nil
		}, nil)
		root.SetArgs(args)
		require.Error(t, root.Execute())
	}
	called := false
	root := command("v0.1.0", func(_ context.Context, dest string, files map[string][]byte, _ scaffold.Runner) error {
		called = true
		require.Equal(t, "dest", dest)
		require.Contains(t, files, "dome.json")
		return nil
	}, nil)
	root.SetOut(&bytes.Buffer{})
	root.SetArgs([]string{"init", "dest", "--module", "example.com/consumer"})
	require.NoError(t, root.Execute())
	require.True(t, called)
}

func TestVersionAndDevelopmentBuild(t *testing.T) {
	var out bytes.Buffer
	root := command("v0.1.0", nil, nil)
	root.SetOut(&out)
	root.SetArgs([]string{"version"})
	require.NoError(t, root.Execute())
	require.Equal(t, "v0.1.0\n", out.String())
	root = command("dev", nil, nil)
	root.SetArgs([]string{"init", "dest", "--module", "example.com/consumer"})
	require.Error(t, root.Execute())
}
