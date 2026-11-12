//go:build integration

package agent_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/jani123/dome/agent"
	"github.com/jani123/dome/cli"
	"github.com/jani123/dome/config"
	"github.com/jani123/dome/core/resource"
	"github.com/jani123/dome/modules/file"
	"github.com/stretchr/testify/require"
)

func TestStandaloneAgentAndEvaluationFailure(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "managed")
	fail := false
	desired := func(context.Context, config.Document) ([]resource.Resource, error) {
		if fail {
			return nil, errors.New("invalid configuration")
		}
		return []resource.Resource{&file.File{Path: path, Content: []byte("value"), Mode: 0o600}}, nil
	}
	a, err := agent.New(agent.WithModules(file.Module{}), agent.WithSQLite(filepath.Join(dir, "state.db")), agent.WithDesiredState(desired))
	require.NoError(t, err)
	p, err := a.Apply(context.Background(), cli.Options{})
	require.NoError(t, err)
	require.Equal(t, resource.Create, p.Changes[0].Action)
	a, err = agent.New(agent.WithModules(file.Module{}), agent.WithSQLite(filepath.Join(dir, "state.db")), agent.WithDesiredState(desired))
	require.NoError(t, err)
	p, err = a.Apply(context.Background(), cli.Options{})
	require.NoError(t, err)
	require.Equal(t, resource.Noop, p.Changes[0].Action)
	fail = true
	_, err = a.Apply(context.Background(), cli.Options{})
	require.Error(t, err)
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "value", string(b))
	status, err := a.Status(context.Background(), cli.Options{})
	require.NoError(t, err)
	require.NotNil(t, status)
}
