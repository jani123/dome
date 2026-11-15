// Command dome initializes version-matched consumer and resource-module projects.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/jani123/dome/cli"
	"github.com/jani123/dome/core/fault"
	"github.com/jani123/dome/internal/scaffold"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

// version is set by tagged release builds. go run/install uses the resolved
// module build metadata; unversioned local development builds remain dev.
var version = "dev"

func buildVersion() string {
	info, _ := debug.ReadBuildInfo()
	return resolveBuildVersion(version, info)
}

func resolveBuildVersion(explicit string, info *debug.BuildInfo) string {
	if explicit != "dev" {
		return explicit
	}
	if info != nil && semver.IsValid(info.Main.Version) && semver.Canonical(info.Main.Version) == info.Main.Version {
		return info.Main.Version
	}
	return explicit
}

type publisher func(context.Context, string, map[string][]byte, scaffold.Runner) error

type modulePublisher func(context.Context, string, map[string]scaffold.File, scaffold.Runner) error

func command(v string, publish publisher, resolve scaffold.Runner) *cobra.Command {
	return commandWithModule(v, publish, resolve, scaffold.PublishModule, scaffold.PrepareModule)
}

func commandWithModule(v string, publish publisher, resolve scaffold.Runner, modulePublish modulePublisher, modulePrepare scaffold.Runner) *cobra.Command {
	root := &cobra.Command{Use: "dome", Short: "Initialize Go infrastructure agents and resource modules", SilenceUsage: true, SilenceErrors: true}
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error { return fault.Wrap(fault.Validation, err) })
	var modulePath string
	var external []string
	init := &cobra.Command{Use: "init <directory> --project-module <module-path>", Short: "Create a fresh version-pinned project (never applies resources)", Args: func(cmd *cobra.Command, args []string) error {
		return fault.Wrap(fault.Validation, cobra.ExactArgs(1)(cmd, args))
	}, RunE: func(cmd *cobra.Command, args []string) error {
		if modulePath == "" {
			return fault.New(fault.Validation, "--project-module is required: name the Go module for the project being created")
		}
		files, err := scaffold.RenderWithModules(modulePath, v, external)
		if err != nil {
			return err
		}
		prepare := resolve
		if len(external) > 0 {
			prepare = scaffold.WithPinnedDependencies(resolve, files["go.mod"])
		}
		if err := publish(cmd.Context(), args[0], files, prepare); err != nil {
			return err
		}
		if len(external) > 0 {
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), "External modules registered as Module{}; configure required parameters in cmd/agent/main.go before building."); err != nil {
				return err
			}
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout(), "Created project. Open it in Codex; see README.md for build and plan commands.")
		return err
	}}
	init.Flags().StringVar(&modulePath, "project-module", "", "Go module identity of the project being created (required; need not be hosted)")
	init.Flags().StringArrayVar(&external, "module", nil, "External root Go module to import and register as Module{}: path@version (repeatable; configure parameters in generated Go)")
	// Resource modules retain their existing flag and separate storage.
	var resourceModulePath string
	var resourceName string
	moduleCmd := &cobra.Command{Use: "module", Short: "Develop standalone resource modules"}
	moduleInit := &cobra.Command{Use: "init --module <go-module-path> --name <name>", Short: "Initialize the current empty directory as a resource-module repository", Args: func(cmd *cobra.Command, args []string) error {
		return fault.Wrap(fault.Validation, cobra.NoArgs(cmd, args))
	}, RunE: func(cmd *cobra.Command, _ []string) error {
		files, err := scaffold.RenderModule(resourceModulePath, resourceName, v)
		if err != nil {
			return err
		}
		if err := modulePublish(cmd.Context(), ".", files, modulePrepare); err != nil {
			return err
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout(), "Created module repository. Run task dev:setup and task check; see README.md.")
		return err
	}}
	moduleInit.Flags().StringVar(&resourceModulePath, "module", "", "External Go module path (required)")
	moduleInit.Flags().StringVar(&resourceName, "name", "", "Importable lowercase package and resource identity (required)")
	moduleCmd.AddCommand(moduleInit)
	root.AddCommand(moduleCmd)
	root.AddCommand(init, &cobra.Command{Use: "version", Args: func(cmd *cobra.Command, args []string) error {
		return fault.Wrap(fault.Validation, cobra.NoArgs(cmd, args))
	}, RunE: func(cmd *cobra.Command, _ []string) error { _, err := fmt.Fprintln(cmd.OutOrStdout(), v); return err }})
	return root
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if err := command(buildVersion(), scaffold.Publish, scaffold.Tidy).ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "dome:", fault.Category(err), err)
		os.Exit(cli.ExitCode(err))
	}
}
