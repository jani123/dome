package agent_test

import (
	"context"
	"testing"

	"github.com/jani123/dome/agent"
	"github.com/jani123/dome/cli"
	"github.com/jani123/dome/modules/file"
	"github.com/stretchr/testify/require"
)

func TestAgentOptionsAndExplicitRegistration(t *testing.T) {
	_, err := agent.New(agent.WithModules(file.Module{}, file.Module{}))
	require.Error(t, err)
	_, err = agent.New(agent.WithConcurrency(0))
	require.Error(t, err)
	a, err := agent.New(agent.WithVersion("test"))
	require.NoError(t, err)
	require.Equal(t, "test", a.Version())
	_, err = a.Apply(context.Background(), cli.Options{})
	require.ErrorContains(t, err, "desired-state")
	require.Empty(t, a.Modules())
	require.Empty(t, a.Playbooks())
}
