package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jani123/dome/core/fault"
	"github.com/jani123/dome/workspace"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// CommandOption injects CLI boundaries without extending Backend.
type (
	CommandOption   func(*commandSettings)
	commandSettings struct {
		store    workspace.Store
		terminal func(io.Reader) bool
	}
)

// WithWorkspaceStore replaces filesystem I/O for command construction and tests.
func WithWorkspaceStore(s workspace.Store) CommandOption {
	return func(c *commandSettings) { c.store = s }
}

// WithTerminalDetector replaces interactive input detection for command tests.
func WithTerminalDetector(fn func(io.Reader) bool) CommandOption {
	return func(c *commandSettings) { c.terminal = fn }
}

func defaultSettings() commandSettings {
	return commandSettings{store: workspace.OSStore{}, terminal: func(r io.Reader) bool { f, ok := r.(*os.File); return ok && term.IsTerminal(int(f.Fd())) }}
}

var setupFlags = []string{"state", "config", "secrets", "git-url", "git-ref", "git-path", "git-allow-http", "ssh-key", "known-hosts"}

func bindSetup(cmd *cobra.Command, o *Options, hidden bool) {
	f := cmd.Flags()
	if hidden {
		f = cmd.PersistentFlags()
	}
	f.StringVar(&o.StatePath, "state", "", "SQLite state database path")
	f.StringVar(&o.LocalPath, "config", "", "Local typed YAML configuration file")
	f.StringVar(&o.SecretsPath, "secrets", "", "SOPS YAML secrets file (repository-relative with Git)")
	f.StringVar(&o.Git.URL, "git-url", "", "Configuration repository URL")
	f.BoolVar(&o.Git.AllowHTTP, "git-allow-http", false, "Explicitly allow unencrypted HTTP retrieval")
	f.StringVar(&o.Git.Ref, "git-ref", "HEAD", "Configuration commit, tag or branch")
	f.StringVar(&o.Git.Path, "git-path", "config.yaml", "Repository-relative configuration path")
	f.StringVar(&o.Git.SSHKey, "ssh-key", "", "SSH private-key file reference")
	f.StringVar(&o.Git.KnownHosts, "known-hosts", "", "Verified SSH known_hosts file reference")
	if hidden {
		for _, name := range setupFlags {
			_ = f.MarkHidden(name)
		}
	}
}

func changed(cmd *cobra.Command, name string) bool {
	if f := cmd.Flags().Lookup(name); f != nil && f.Changed {
		return true
	}
	for parent := cmd; parent != nil; parent = parent.Parent() {
		if f := parent.PersistentFlags().Lookup(name); f != nil && f.Changed {
			return true
		}
	}
	return false
}

func profileFromFlags(cmd *cobra.Command, p workspace.Profile, o Options) workspace.Profile {
	// Selecting a source replaces source-specific settings as a unit, including
	// secrets. Refinement flags without a source selection retain the saved source.
	if changed(cmd, "config") {
		p = workspace.Profile{SchemaVersion: p.SchemaVersion, State: p.State, Source: "local", Config: o.LocalPath}
	}
	if changed(cmd, "git-url") {
		p = workspace.Profile{SchemaVersion: p.SchemaVersion, State: p.State, Source: "git", GitURL: o.Git.URL}
	}
	for _, v := range []struct {
		name  string
		dest  *string
		value string
	}{
		{"state", &p.State, o.StatePath},
		{"secrets", &p.Secrets, o.SecretsPath},
		{"git-ref", &p.GitRef, o.Git.Ref},
		{"git-path", &p.GitPath, o.Git.Path},
		{"ssh-key", &p.SSHKey, o.Git.SSHKey},
		{"known-hosts", &p.KnownHosts, o.Git.KnownHosts},
	} {
		if changed(cmd, v.name) {
			*v.dest = v.value
		}
	}
	if changed(cmd, "git-allow-http") {
		p.AllowHTTP = o.Git.AllowHTTP
	}
	return p
}

func profileOptions(p workspace.Profile, json bool) Options {
	return Options{StatePath: p.State, LocalPath: p.Config, SecretsPath: p.Secrets, Git: p.GitOptions(), JSON: json}
}

func runtimeSettings(cmd *cobra.Command) bool {
	switch cmd.Name() {
	case "plan", "apply", "run", "status", "inspect":
		return true
	case "list":
		return cmd.Parent() != nil && cmd.Parent().Name() == "operations"
	}
	return false
}

func attachWorkspace(root *cobra.Command, o *Options, c commandSettings) {
	var dir, profilePath string
	root.PersistentFlags().StringVar(&dir, "workspace", "", "Runtime workspace directory (default: current directory)")
	root.PersistentFlags().StringVar(&profilePath, "agent-config", "", "Explicit runtime profile (default: <workspace>/.dome/agent.yaml)")
	root.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		if changed(cmd, "config") && changed(cmd, "git-url") {
			return fault.New(fault.Validation, "choose either --config or --git-url")
		}
		if !runtimeSettings(cmd) {
			return nil
		}
		absolute, err := filepath.Abs(dir)
		if err != nil {
			return fault.Wrap(fault.Configuration, err)
		}
		p, found, err := c.store.Load(cmd.Context(), absolute, profilePath)
		if err != nil {
			return err
		}
		if !found {
			if profilePath != "" {
				return fault.New(fault.Configuration, "explicit agent configuration was not found")
			}
			return nil // Existing Go defaults and legacy CLI source behavior.
		}
		p = p.Defaults()
		if err = p.Validate(); err != nil {
			return err
		}
		for _, name := range []string{"state", "git-ref", "git-path"} {
			if changed(cmd, name) {
				value, _ := cmd.Flags().GetString(name)
				if value == "" {
					return fault.New(fault.Validation, "explicit runtime setting cannot be empty: "+name)
				}
			}
		}
		p = profileFromFlags(cmd, p, *o).Defaults()
		if err = p.Validate(); err != nil {
			return err
		}
		*o = profileOptions(p.Resolve(absolute), o.JSON)
		return nil
	}
	var noninteractive bool
	initOptions := Options{}
	init := &cobra.Command{Use: "init", Short: "Initialize host workspace and save runtime configuration", Args: checkedArgs(cobra.NoArgs), RunE: func(cmd *cobra.Command, _ []string) error {
		absolute, err := filepath.Abs(dir)
		if err != nil {
			return fault.Wrap(fault.Configuration, err)
		}
		p, found, err := c.store.Load(cmd.Context(), absolute, profilePath)
		if err != nil {
			return err
		}
		if !found && profilePath != "" {
			return fault.New(fault.Configuration, "explicit agent configuration was not found")
		}
		if found {
			p = p.Defaults()
			if err = p.Validate(); err != nil {
				return err
			}
		} else {
			p.SchemaVersion = 1
		}
		// Visible init-local flags shadow hidden persistent flags. Root-position
		// legacy flags are still accepted and take effect unless shadowed explicitly.
		supplied := *o
		for _, name := range setupFlags {
			if f := cmd.Flags().Lookup(name); f != nil && f.Changed {
				switch name {
				case "state":
					supplied.StatePath = initOptions.StatePath
				case "config":
					supplied.LocalPath = initOptions.LocalPath
				case "secrets":
					supplied.SecretsPath = initOptions.SecretsPath
				case "git-url":
					supplied.Git.URL = initOptions.Git.URL
				case "git-ref":
					supplied.Git.Ref = initOptions.Git.Ref
				case "git-path":
					supplied.Git.Path = initOptions.Git.Path
				case "ssh-key":
					supplied.Git.SSHKey = initOptions.Git.SSHKey
				case "known-hosts":
					supplied.Git.KnownHosts = initOptions.Git.KnownHosts
				case "git-allow-http":
					supplied.Git.AllowHTTP = initOptions.Git.AllowHTTP
				}
			}
		}
		p = profileFromFlags(cmd, p, supplied).Defaults()
		interactive := !noninteractive && !o.JSON && c.terminal(cmd.InOrStdin())
		if err = completeProfile(cmd, promptInput{interactive: interactive, reader: bufio.NewReader(cmd.InOrStdin())}, &p); err != nil {
			return err
		}
		p = p.Defaults()
		if err = p.Validate(); err != nil {
			return err
		}
		resolved := p.Resolve(absolute)
		for _, path := range []string{resolved.Config, resolved.SSHKey, resolved.KnownHosts} {
			if path != "" {
				if err = c.store.CheckFile(cmd.Context(), path); err != nil {
					return err
				}
			}
		}
		if p.Source == "local" && resolved.Secrets != "" {
			if err = c.store.CheckFile(cmd.Context(), resolved.Secrets); err != nil {
				return err
			}
		}
		if err = c.store.Save(cmd.Context(), absolute, p); err != nil {
			return err
		}
		if o.JSON {
			return output(cmd.OutOrStdout(), map[string]string{"workspace": absolute, "profile": filepath.Join(absolute, ".dome/agent.yaml")})
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout(), "Workspace initialized. Run agent plan, then agent apply.")
		return err
	}}
	bindSetup(init, &initOptions, false)
	init.Flags().BoolVar(&noninteractive, "non-interactive", false, "Require complete settings without prompting")
	root.AddCommand(init)
}

type promptInput struct {
	interactive bool
	reader      *bufio.Reader
}

func (in promptInput) ask(ctx context.Context, cmd *cobra.Command, label string) (string, error) {
	if !in.interactive {
		return "", fault.New(fault.Validation, "missing initialization setting: "+label+"; supply flags or --agent-config")
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if _, err := fmt.Fprint(cmd.ErrOrStderr(), label+": "); err != nil {
		return "", err
	}
	// Terminal reads are cancellable without leaving command execution blocked.
	type result struct {
		line string
		err  error
	}
	ch := make(chan result, 1)
	go func() { line, err := in.reader.ReadString('\n'); ch <- result{line, err} }()
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case r := <-ch:
		if r.err != nil {
			return "", fault.New(fault.Validation, "initialization input ended before setup completed")
		}
		return strings.TrimSpace(r.line), nil
	}
}

func completeProfile(cmd *cobra.Command, in promptInput, p *workspace.Profile) error {
	wizard := p.Source == ""
	require := func(value *string, label string) error {
		if *value != "" {
			return nil
		}
		v, err := in.ask(cmd.Context(), cmd, label)
		if err != nil {
			return err
		}
		if v == "" {
			return fault.New(fault.Validation, "required initialization setting is empty: "+label)
		}
		*value = v
		return nil
	}
	optional := func(value *string, label string) error {
		v, err := in.ask(cmd.Context(), cmd, label)
		if err != nil {
			return err
		}
		if v != "" {
			*value = v
		}
		return nil
	}
	if err := require(&p.Source, "Source (local or git)"); err != nil {
		return err
	}
	switch p.Source {
	case "local":
		if err := require(&p.Config, "Local configuration file"); err != nil {
			return err
		}
	case "git":
		if err := require(&p.GitURL, "Git URL"); err != nil {
			return err
		}
		if wizard && !changed(cmd, "git-ref") {
			if err := optional(&p.GitRef, "Git ref [HEAD]"); err != nil {
				return err
			}
		}
		if wizard && !changed(cmd, "git-path") {
			if err := optional(&p.GitPath, "Git configuration path [config.yaml]"); err != nil {
				return err
			}
		}
		if strings.HasPrefix(p.GitURL, "http://") && !p.AllowHTTP && in.interactive && !changed(cmd, "git-allow-http") {
			answer, err := in.ask(cmd.Context(), cmd, "Allow unencrypted HTTP (yes/no) [no]")
			if err != nil {
				return err
			}
			p.AllowHTTP = answer == "yes"
		}
		if strings.HasPrefix(p.GitURL, "ssh://") {
			if err := require(&p.SSHKey, "SSH key file"); err != nil {
				return err
			}
			if err := require(&p.KnownHosts, "Known hosts file"); err != nil {
				return err
			}
		}
	default:
		return fault.New(fault.Validation, "source must be local or git")
	}
	if wizard && !changed(cmd, "state") {
		if err := optional(&p.State, "State database [.dome/state.db]"); err != nil {
			return err
		}
	}
	if wizard && !changed(cmd, "secrets") {
		if err := optional(&p.Secrets, "Encrypted secrets path [none]"); err != nil {
			return err
		}
	}
	return nil
}
