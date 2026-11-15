package scaffold

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jani123/dome/core/fault"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
)

type externalModule struct{ Path, Version, Alias string }

func externalModules(project string, selections []string) ([]externalModule, error) {
	var result []externalModule
	seen := make(map[string]bool)
	for i, selection := range selections {
		path, version, ok := strings.Cut(selection, "@")
		if !ok || module.Check(path, version) != nil || !semver.IsValid(version) || semver.Canonical(version) != version || semver.Build(version) != "" {
			return nil, fault.New(fault.Validation, "--module requires a root Go module path and exact version: module/path@vX.Y.Z")
		}
		if path == project || strings.HasPrefix(path, project+"/") || path == "github.com/jani123/dome" || strings.HasPrefix(path, "github.com/jani123/dome/") || seen[path] {
			return nil, fault.New(fault.Validation, "--module must select distinct external modules, not the project or Dome's built-in modules")
		}
		seen[path] = true
		result = append(result, externalModule{path, version, fmt.Sprintf("external%d", i+1)})
	}
	return result, nil
}

// WithPinnedDependencies checks staged dependency resolution before publication.
// No compilation is required: generated registrations may still need parameters.
func WithPinnedDependencies(next Runner, expected []byte) Runner {
	return func(ctx context.Context, dir string) error {
		if next == nil {
			return fault.New(fault.Configuration, "external modules require dependency resolution")
		}
		if err := next(ctx, dir); err != nil {
			return err
		}
		actual, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err != nil {
			return fault.Wrap(fault.Configuration, err)
		}
		return checkPinnedDependencies(expected, actual)
	}
}

func checkPinnedDependencies(expected, actual []byte) error {
	before, err := modfile.Parse("go.mod", expected, nil)
	if err != nil {
		return fault.Wrap(fault.Configuration, err)
	}
	after, err := modfile.Parse("go.mod", actual, nil)
	if err != nil {
		return fault.Wrap(fault.Configuration, err)
	}
	resolved := make(map[string]string)
	for _, dependency := range after.Require {
		resolved[dependency.Mod.Path] = dependency.Mod.Version
	}
	for _, dependency := range before.Require {
		if resolved[dependency.Mod.Path] != dependency.Mod.Version {
			return fault.New(fault.Conflict, "dependency resolution changed a requested module version; choose compatible module versions", "module", dependency.Mod.Path)
		}
	}
	return nil
}
