package main

import (
	"bytes"
	"context"
	"errors"
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

func TestModuleInitCommand(t *testing.T) {
	for _, args := range [][]string{{"module", "init"}, {"module", "init", "--module", "example.com/demo", "--name", "main"}, {"module", "init", "extra", "--module", "example.com/demo", "--name", "demo"}} {
		root := commandWithModule("v0.1.0", nil, nil, func(context.Context, string, map[string]scaffold.File, scaffold.Runner) error {
			t.Fatal("invalid module request published")
			return nil
		}, nil)
		root.SetArgs(args)
		require.Error(t, root.Execute())
	}
	called := false
	root := commandWithModule("v0.1.0", nil, nil, func(_ context.Context, dest string, files map[string]scaffold.File, _ scaffold.Runner) error {
		called = true
		require.Equal(t, ".", dest)
		require.Contains(t, files, "Taskfile.yml")
		return nil
	}, nil)
	root.SetOut(&bytes.Buffer{})
	root.SetArgs([]string{"module", "init", "--module", "example.com/demo", "--name", "demo"})
	require.NoError(t, root.Execute())
	require.True(t, called)
}

func TestModulePublicationFailure(t *testing.T) {
	failure := errors.New("publisher failed")
	root := commandWithModule("v0.1.0", nil, nil, func(context.Context, string, map[string]scaffold.File, scaffold.Runner) error { return failure }, nil)
	root.SetArgs([]string{"module", "init", "--module", "example.com/demo", "--name", "demo"})
	require.ErrorIs(t, root.Execute(), failure)
}
