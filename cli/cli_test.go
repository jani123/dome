package cli_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/jani123/dome/cli"
	"github.com/jani123/dome/core/fault"
	"github.com/jani123/dome/core/reconcile"
	"github.com/jani123/dome/core/state"
	"github.com/stretchr/testify/require"
)

type backend struct{}

func (backend) Plan(context.Context, cli.Options) (*reconcile.Plan, error) {
	return &reconcile.Plan{Revision: "abc", Changes: []reconcile.Change{}}, nil
}

func (b backend) Apply(ctx context.Context, o cli.Options) (*reconcile.Plan, error) {
	return b.Plan(ctx, o)
}

func (backend) Status(context.Context, cli.Options) (any, error) {
	return map[string]string{"revision": "abc"}, nil
}
func (backend) Modules() []string   { return []string{"file"} }
func (backend) Playbooks() []string { return []string{"demo"} }
func (backend) RunPlaybook(context.Context, cli.Options, string, string) (string, error) {
	return "execution", nil
}

func (backend) Operations(context.Context, cli.Options) ([]state.Operation, error) {
	return []state.Operation{{ID: "operation", Status: "succeeded"}}, nil
}
func (backend) Version() string { return "test-version" }
func TestCommands(t *testing.T) {
	for _, c := range []struct {
		args []string
		want string
	}{{[]string{"plan", "--json"}, "abc"}, {[]string{"apply", "--json"}, "abc"}, {[]string{"run", "--json"}, "abc"}, {[]string{"status"}, "revision"}, {[]string{"modules"}, "file"}, {[]string{"playbook", "list"}, "demo"}, {[]string{"playbook", "run", "demo"}, "execution"}, {[]string{"operations", "list"}, "operation"}, {[]string{"operations", "inspect", "operation"}, "succeeded"}, {[]string{"version"}, "test-version"}, {[]string{"--help"}, "Reconcile"}} {
		t.Run(c.args[0]+c.want, func(t *testing.T) {
			root := cli.New(backend{})
			var out bytes.Buffer
			root.SetOut(&out)
			root.SetArgs(c.args)
			require.NoError(t, root.Execute())
			require.Contains(t, out.String(), c.want)
		})
	}
}

func TestInvalidSourceAndExitCodes(t *testing.T) {
	root := cli.New(backend{})
	root.SetArgs([]string{"plan", "--config", "local", "--git-url", "ssh://host/x"})
	err := root.Execute()
	require.Error(t, err)
	require.Equal(t, 2, cli.ExitCode(err))
	require.Equal(t, 1, cli.ExitCode(fault.New(fault.Apply, "failed")))
	require.Zero(t, cli.ExitCode(nil))
	root = cli.New(backend{})
	root.SetArgs([]string{"operations", "inspect", "missing"})
	require.Error(t, root.Execute())
}

func TestInvalidCommandsUseValidationExitCode(t *testing.T) {
	for _, args := range [][]string{{"not-a-command"}, {"playbook", "not-a-command"}, {"operations", "not-a-command"}, {"playbook", "run"}, {"apply", "--not-a-flag"}} {
		t.Run(args[0], func(t *testing.T) {
			root := cli.New(backend{})
			root.SetArgs(args)
			err := root.Execute()
			require.Error(t, err)
			require.Equal(t, 2, cli.ExitCode(err))
		})
	}
}
