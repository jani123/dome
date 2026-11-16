// Package cli exposes the agent's Cobra commands without importing concrete modules.
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/jani123/dome/config"
	"github.com/jani123/dome/core/dag"
	"github.com/jani123/dome/core/fault"
	"github.com/jani123/dome/core/reconcile"
	"github.com/jani123/dome/core/state"
	"github.com/spf13/cobra"
)

type (
	// Options overrides configured state and sources for a single invocation.
	Options struct {
		StatePath   string
		LocalPath   string
		SecretsPath string
		Git         config.GitOptions
		JSON        bool
	}
	// Backend is the injectable runtime contract used by the consumer CLI.
	Backend interface {
		Plan(context.Context, Options) (*reconcile.Plan, error)
		Apply(context.Context, Options) (*reconcile.Plan, error)
		Status(context.Context, Options) (any, error)
		Modules() []string
		Playbooks() []string
		RunPlaybook(context.Context, Options, string, string) (string, error)
		Operations(context.Context, Options) ([]state.Operation, error)
		Version() string
	}
)

func output(w io.Writer, v any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

type clock interface {
	Sleep(context.Context, time.Duration) error
}

// New constructs commands for an injectable backend without running them.
func New(b Backend) *cobra.Command { return NewWithOptions(b) }

// NewWithOptions constructs commands with explicitly injected CLI boundaries.
func NewWithOptions(b Backend, options ...CommandOption) *cobra.Command {
	return newWithClock(b, dag.RealClock{}, options...)
}

func newWithClock(b Backend, clock clock, options ...CommandOption) *cobra.Command {
	settings := defaultSettings()
	for _, option := range options {
		if option != nil {
			option(&settings)
		}
	}
	var o Options
	root := &cobra.Command{Use: "agent", Short: "Reconcile infrastructure and execute durable playbooks", SilenceErrors: true, SilenceUsage: true, Args: groupArgs, RunE: groupHelp}
	flags := root.PersistentFlags()
	bindSetup(root, &o, true)
	flags.BoolVar(&o.JSON, "json", false, "Print machine-readable JSON")
	root.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error { return fault.Wrap(fault.Validation, err) })
	attachWorkspace(root, &o, settings)
	showPlan := func(cmd *cobra.Command, p *reconcile.Plan) error {
		if o.JSON {
			return output(cmd.OutOrStdout(), p)
		}
		for _, c := range p.Changes {
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", c.Action, c.ID.String()); err != nil {
				return err
			}
		}
		return nil
	}
	root.AddCommand(&cobra.Command{Use: "plan", Short: "Observe resources and describe changes without applying them", Args: checkedArgs(cobra.NoArgs), RunE: func(cmd *cobra.Command, args []string) error {
		p, err := b.Plan(cmd.Context(), o)
		if err != nil {
			return err
		}
		return showPlan(cmd, p)
	}})
	root.AddCommand(&cobra.Command{Use: "apply", Short: "Apply and verify desired state, then prune owned resources", Args: checkedArgs(cobra.NoArgs), RunE: func(cmd *cobra.Command, args []string) error {
		p, err := b.Apply(cmd.Context(), o)
		if err != nil {
			return err
		}
		return showPlan(cmd, p)
	}})
	var interval time.Duration
	run := &cobra.Command{Use: "run", Short: "Continuously reconcile in the foreground", Args: checkedArgs(cobra.NoArgs), RunE: func(cmd *cobra.Command, args []string) error {
		if interval <= 0 {
			return fault.New(fault.Validation, "interval must be positive")
		}
		for {
			p, err := b.Apply(cmd.Context(), o)
			if err != nil {
				return err
			}
			if err = showPlan(cmd, p); err != nil {
				return err
			}
			if err = clock.Sleep(cmd.Context(), interval); err != nil {
				return err
			}
		}
	}}
	run.Flags().DurationVar(&interval, "interval", 30*time.Second, "Wait between reconciliations (must be positive)")
	root.AddCommand(run)
	root.AddCommand(&cobra.Command{Use: "status", Short: "Show managed resources, observations, and applied revision", Args: checkedArgs(cobra.NoArgs), RunE: func(cmd *cobra.Command, args []string) error {
		v, err := b.Status(cmd.Context(), o)
		if err != nil {
			return err
		}
		return output(cmd.OutOrStdout(), v)
	}})
	root.AddCommand(&cobra.Command{Use: "modules", Short: "List explicitly registered modules", Args: checkedArgs(cobra.NoArgs), RunE: func(cmd *cobra.Command, args []string) error { return output(cmd.OutOrStdout(), b.Modules()) }})
	book := &cobra.Command{Use: "playbook", Short: "Inspect and execute registered DAG playbooks", Args: groupArgs, RunE: groupHelp}
	book.AddCommand(&cobra.Command{Use: "list", Short: "List registered playbooks", Args: checkedArgs(cobra.NoArgs), RunE: func(cmd *cobra.Command, args []string) error { return output(cmd.OutOrStdout(), b.Playbooks()) }})
	var resume string
	bookRun := &cobra.Command{Use: "run <name>", Short: "Execute a playbook or resume an execution after observing actual state", Args: checkedArgs(cobra.ExactArgs(1)), RunE: func(cmd *cobra.Command, args []string) error {
		id, err := b.RunPlaybook(cmd.Context(), o, args[0], resume)
		if id != "" {
			if printErr := output(cmd.OutOrStdout(), map[string]string{"execution_id": id}); printErr != nil {
				return printErr
			}
		}
		return err
	}}
	bookRun.Flags().StringVar(&resume, "resume", "", "Execution ID to resume")
	book.AddCommand(bookRun)
	root.AddCommand(book)
	operations := &cobra.Command{Use: "operations", Short: "Inspect durable resource and playbook operation history", Args: groupArgs, RunE: groupHelp}
	operations.AddCommand(&cobra.Command{Use: "list", Short: "List resource operations and DAG executions", Args: checkedArgs(cobra.NoArgs), RunE: func(cmd *cobra.Command, args []string) error {
		v, err := b.Operations(cmd.Context(), o)
		if err != nil {
			return err
		}
		return output(cmd.OutOrStdout(), v)
	}})
	operations.AddCommand(&cobra.Command{Use: "inspect <id>", Short: "Inspect an operation by stable ID", Args: checkedArgs(cobra.ExactArgs(1)), RunE: func(cmd *cobra.Command, args []string) error {
		v, err := b.Operations(cmd.Context(), o)
		if err != nil {
			return err
		}
		for _, operation := range v {
			if operation.ID == args[0] {
				return output(cmd.OutOrStdout(), operation)
			}
		}
		return fault.New(fault.Validation, "operation was not found", "operation_id", args[0])
	}})
	root.AddCommand(operations)
	root.AddCommand(&cobra.Command{Use: "version", Short: "Print the agent version", Args: checkedArgs(cobra.NoArgs), RunE: func(cmd *cobra.Command, args []string) error {
		_, err := fmt.Fprintln(cmd.OutOrStdout(), b.Version())
		return err
	}})
	return root
}

// ExitCode distinguishes invalid input from execution failures.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	switch fault.Category(err) {
	case string(fault.Validation), string(fault.Configuration):
		return 2
	default:
		return 1
	}
}

func checkedArgs(validate cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		return fault.Wrap(fault.Validation, validate(cmd, args))
	}
}

func groupArgs(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return fault.New(fault.Validation, fmt.Sprintf("unknown command %q for %q", args[0], cmd.CommandPath()))
	}
	return nil
}
func groupHelp(cmd *cobra.Command, args []string) error { return cmd.Help() }
