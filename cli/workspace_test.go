package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/jani123/dome/core/reconcile"
	"github.com/jani123/dome/core/state"
	"github.com/jani123/dome/workspace"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

type memoryWorkspace struct {
	saveErr  error
	checkErr error
	profile  workspace.Profile
	found    bool
	err      error
	saves    int
	checked  []string
}

func (s *memoryWorkspace) Load(context.Context, string, string) (workspace.Profile, bool, error) {
	return s.profile, s.found, s.err
}

func (s *memoryWorkspace) Save(_ context.Context, _ string, p workspace.Profile) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.profile = p
	s.found = true
	s.saves++
	return nil
}

func (s *memoryWorkspace) CheckFile(_ context.Context, p string) error {
	s.checked = append(s.checked, p)
	return s.checkErr
}

type capturingBackend struct{ options []Options }

func (b *capturingBackend) Plan(_ context.Context, o Options) (*reconcile.Plan, error) {
	b.options = append(b.options, o)
	return &reconcile.Plan{}, nil
}

func (b *capturingBackend) Apply(c context.Context, o Options) (*reconcile.Plan, error) {
	return b.Plan(c, o)
}

func (b *capturingBackend) Status(c context.Context, o Options) (any, error) {
	_, err := b.Plan(c, o)
	return map[string]string{}, err
}

func (b *capturingBackend) Operations(c context.Context, o Options) ([]state.Operation, error) {
	_, err := b.Plan(c, o)
	return []state.Operation{{ID: "op"}}, err
}

func (b *capturingBackend) RunPlaybook(c context.Context, o Options, _, _ string) (string, error) {
	_, err := b.Plan(c, o)
	return "id", err
}
func (*capturingBackend) Modules() []string   { return nil }
func (*capturingBackend) Playbooks() []string { return nil }
func (*capturingBackend) Version() string     { return "test" }

func TestWorkspaceInitInput(t *testing.T) {
	for _, tt := range []struct {
		name     string
		args     []string
		input    string
		terminal bool
		wantErr  bool
	}{
		{"flags", []string{"init", "--git-url", "ssh://git@example.com/x", "--ssh-key", "key", "--known-hosts", "hosts", "--non-interactive"}, "", false, false},
		{"interactive", []string{"init"}, "git\nssh://git@example.com/x\nmain\nconfig.yaml\nkey\nhosts\n\n\n", true, false},
		{"local", []string{"init", "--config", "config.yaml"}, "", true, false},
		{"missing", []string{"init", "--non-interactive"}, "", false, true},
		{"json", []string{"init", "--json"}, "", true, true},
		{"pipe", []string{"init"}, "git\n", false, true},
		{"eof", []string{"init"}, "git\n", true, true},
		{"http", []string{"init", "--git-url", "http://example.com/x", "--non-interactive"}, "", false, true},
		{"rootflags", []string{"--git-url", "https://example.com/x", "init", "--non-interactive"}, "", false, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s := &memoryWorkspace{}
			b := &capturingBackend{}
			var out bytes.Buffer
			cmd := NewWithOptions(b, WithWorkspaceStore(s), WithTerminalDetector(func(io.Reader) bool { return tt.terminal }))
			cmd.SetIn(strings.NewReader(tt.input))
			cmd.SetOut(&out)
			cmd.SetErr(io.Discard)
			cmd.SetArgs(append([]string{"--workspace", "/workspace"}, tt.args...))
			err := cmd.Execute()
			if tt.wantErr {
				require.Error(t, err)
				require.Zero(t, s.saves)
			} else {
				require.NoError(t, err)
				require.Equal(t, 1, s.saves)
				require.Equal(t, 1, s.profile.SchemaVersion)
			}
			require.Empty(t, b.options)
		})
	}
}

func TestSavedOptionsAndOverrides(t *testing.T) {
	for _, tt := range []struct {
		args  []string
		check func(*testing.T, Options)
	}{
		{[]string{"plan"}, func(t *testing.T, o Options) {
			require.Equal(t, "main", o.Git.Ref)
			require.Equal(t, "/workspace/key", o.Git.SSHKey)
			require.Equal(t, "secrets.yaml", o.SecretsPath)
		}},
		{[]string{"apply", "--config", "local.yaml"}, func(t *testing.T, o Options) {
			require.Equal(t, "/workspace/local.yaml", o.LocalPath)
			require.Empty(t, o.Git.URL)
			require.Empty(t, o.SecretsPath)
		}},
		{[]string{"run", "--git-ref", "other"}, func(t *testing.T, o Options) { require.Equal(t, "other", o.Git.Ref) }},
		{[]string{"status"}, nil},
		{[]string{"operations", "list"}, nil},
		{[]string{"operations", "inspect", "op"}, nil},
		{[]string{"playbook", "run", "test", "--resume", "id"}, nil},
	} {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			p := workspace.Profile{SchemaVersion: 1, Source: "git", State: "state.db", GitURL: "ssh://git@example.com/x", GitRef: "main", GitPath: "values.yaml", SSHKey: "key", KnownHosts: "hosts", Secrets: "secrets.yaml"}
			s := &memoryWorkspace{found: true, profile: p}
			b := &capturingBackend{}
			cmd := NewWithOptions(b, WithWorkspaceStore(s))
			cmd.SetOut(io.Discard)
			cmd.SetArgs(append([]string{"--workspace", "/workspace"}, tt.args...))
			require.NoError(t, cmd.Execute())
			require.Len(t, b.options, 1)
			require.Equal(t, "/workspace/state.db", b.options[0].StatePath)
			if tt.check != nil {
				tt.check(t, b.options[0])
			}
			require.Equal(t, p, s.profile)
			require.Zero(t, s.saves)
		})
	}
}

func TestProfileFailureBlocksRuntimeButNotMetadata(t *testing.T) {
	for _, args := range [][]string{{"apply"}, {"plan"}, {"status"}, {"version"}, {"modules"}, {"playbook", "list"}, {"apply", "--help"}} {
		s := &memoryWorkspace{err: errors.New("cannot load profile")}
		b := &capturingBackend{}
		cmd := NewWithOptions(b, WithWorkspaceStore(s))
		cmd.SetOut(io.Discard)
		cmd.SetArgs(args)
		err := cmd.Execute()
		if len(args) == 1 && args[0] != "version" && args[0] != "modules" {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
		require.Empty(t, b.options)
	}
}

func TestSetupFlagsVisibleOnlyOnInit(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {"init", "--help"}, {"apply", "--help"}, {"plan", "--help"}, {"run", "--help"}, {"status", "--help"}, {"playbook", "run", "--help"}, {"operations", "inspect", "--help"}} {
		var out bytes.Buffer
		cmd := NewWithOptions(&capturingBackend{}, WithWorkspaceStore(&memoryWorkspace{}))
		cmd.SetOut(&out)
		cmd.SetArgs(args)
		require.NoError(t, cmd.Execute())
		for _, name := range setupFlags {
			if args[0] == "init" {
				require.Contains(t, out.String(), "--"+name)
			} else {
				require.NotContains(t, out.String(), "--"+name)
			}
		}
		require.Contains(t, out.String(), "--workspace")
	}
}

func TestInitCancellationAndFalseOverride(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	s := &memoryWorkspace{}
	cmd := NewWithOptions(&capturingBackend{}, WithWorkspaceStore(s), WithTerminalDetector(func(io.Reader) bool { return true }))
	cmd.SetArgs([]string{"init"})
	require.ErrorIs(t, cmd.ExecuteContext(ctx), context.Canceled)
	require.Zero(t, s.saves)
	p := workspace.Profile{SchemaVersion: 1, Source: "git", GitURL: "http://example.com/x", AllowHTTP: true}.Defaults()
	s = &memoryWorkspace{profile: p, found: true}
	b := &capturingBackend{}
	cmd = NewWithOptions(b, WithWorkspaceStore(s))
	cmd.SetArgs([]string{"apply", "--git-allow-http=false"})
	require.Error(t, cmd.Execute())
	require.Empty(t, b.options)
}

func TestInitStoreFailuresDoNotExecuteBackend(t *testing.T) {
	for _, stage := range []string{"load", "reference", "publication"} {
		t.Run(stage, func(t *testing.T) {
			failure := errors.New("injected boundary failure")
			s := &memoryWorkspace{}
			switch stage {
			case "load":
				s.err = failure
			case "reference":
				s.checkErr = failure
			case "publication":
				s.saveErr = failure
			}
			b := &capturingBackend{}
			cmd := NewWithOptions(b, WithWorkspaceStore(s))
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)
			cmd.SetArgs([]string{"init", "--config", "config.yaml", "--non-interactive"})
			require.ErrorIs(t, cmd.Execute(), failure)
			require.Zero(t, s.saves)
			require.Empty(t, b.options)
		})
	}
}

func TestHelpReusedCommandTree(t *testing.T) {
	var out bytes.Buffer
	cmd := NewWithOptions(&capturingBackend{}, WithWorkspaceStore(&memoryWorkspace{}))
	cmd.SetOut(&out)
	for _, name := range []string{"init", "apply", "plan", "init", "status"} {
		out.Reset()
		cmd.SetArgs([]string{name, "--help"})
		require.NoError(t, cmd.Execute())
		if name == "init" {
			require.Contains(t, out.String(), "--git-url")
		} else {
			require.NotContains(t, out.String(), "--git-url")
		}
	}
}

func TestExplicitProfileAndEmptyOverrides(t *testing.T) {
	s := &memoryWorkspace{}
	b := &capturingBackend{}
	cmd := NewWithOptions(b, WithWorkspaceStore(s))
	cmd.SetArgs([]string{"apply", "--agent-config", "missing.yaml"})
	require.Error(t, cmd.Execute())
	require.Empty(t, b.options)
	p := workspace.Profile{SchemaVersion: 1, Source: "local", Config: "config.yaml", State: "existing.db"}
	s = &memoryWorkspace{profile: p, found: true}
	cmd = NewWithOptions(b, WithWorkspaceStore(s))
	cmd.SetArgs([]string{"apply", "--state="})
	require.Error(t, cmd.Execute())
	require.Empty(t, b.options)
}

func TestWizardPreservesExplicitPartialInput(t *testing.T) {
	for _, secrets := range []string{"secrets.enc.yaml", ""} {
		s := &memoryWorkspace{}
		var diagnostics bytes.Buffer
		cmd := NewWithOptions(&capturingBackend{}, WithWorkspaceStore(s), WithTerminalDetector(func(io.Reader) bool { return true }))
		cmd.SetOut(io.Discard)
		cmd.SetErr(&diagnostics)
		cmd.SetIn(strings.NewReader("git\nssh://git@example.com/x\nkey\nhosts\n"))
		cmd.SetArgs([]string{"init", "--git-ref", "main", "--git-path", "values.yaml", "--state", "existing.db", "--secrets", secrets})
		require.NoError(t, cmd.Execute())
		require.Equal(t, "main", s.profile.GitRef)
		require.Equal(t, "values.yaml", s.profile.GitPath)
		require.Equal(t, "existing.db", s.profile.State)
		require.Equal(t, secrets, s.profile.Secrets)
		for _, label := range []string{"Git ref", "Git configuration path", "State database", "Encrypted secrets"} {
			require.NotContains(t, diagnostics.String(), label)
		}
	}
}

// Preserve callers that hold the original CLI constructor as a function value.
var _ func(Backend) *cobra.Command = New
