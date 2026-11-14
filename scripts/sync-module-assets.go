//go:build ignore

// Sync shared developer assets into the embedded module starter. Run at repo root.
package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	pairs := map[string]string{"LICENSE": "LICENSE", ".gitignore": "gitignore", ".openspec-version": "openspec-version", ".hugo-version": "hugo-version", "devbox.json": "devbox.json", "devbox.lock": "devbox.lock", ".agents/OPENSPEC_LICENSE": "_agents/OPENSPEC_LICENSE"}
	for _, name := range []string{"install-openspec.sh", "openspec.sh", "setup-specs.sh", "check-maintainer-skills.go", "check-knowledge.go", "install-hugo.sh", "serve-docs.sh"} {
		pairs["scripts/"+name] = "scripts/" + name
	}
	for _, name := range []string{"validate.go", "workflows.go", "validate_test.go", "workflows_test.go"} {
		pairs["internal/knowledge/"+name] = "internal/knowledge/" + name
	}
	for _, name := range []string{"explore", "propose", "apply", "update", "sync", "archive", "verify"} {
		key := ".agents/skills/openspec-"
		suffix := name + "-change"
		if name == "explore" || name == "propose" {
			suffix = name
		}
		if name == "sync" {
			suffix = "sync-specs"
		}
		pairs[key+suffix+"/SKILL.md"] = "_agents/skills/openspec-" + suffix + "/SKILL.md"
	}
	err := filepath.WalkDir(".devcontainer", func(path string, e os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !e.IsDir() && path != ".devcontainer/README.md" {
			pairs[path] = strings.Replace(path, ".devcontainer/", "_devcontainer/", 1)
		}
		return nil
	})
	if err != nil {
		return err
	}
	check := len(os.Args) > 1 && os.Args[1] == "--check"
	for src, dst := range pairs {
		b, err := os.ReadFile(src)
		if err != nil {
			return err
		}
		s := strings.NewReplacer("github.com/jani123/dome/internal/knowledge", "example.com/__MODULE__/internal/knowledge", "/workspaces/dome", "/workspaces/module", "dome-devbox-", "module-devbox-", "dome-cache-", "module-cache-", "dome-go-", "module-go-", "dome-npm-", "module-npm-", "Dome development", "Module development", "visible_hostname dome-egress", "visible_hostname __NAME__-egress").Replace(string(b))
		b = []byte(s)
		if strings.HasSuffix(dst, ".go") {
			dst += ".tmpl"
			b, err = format.Source(b)
			if err != nil {
				return err
			}
		}
		path := filepath.Join("internal/scaffold/moduletemplates", dst)
		if check {
			old, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if !bytes.Equal(old, b) {
				return fmt.Errorf("module shared asset drift: %s (run go run scripts/sync-module-assets.go)", src)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, b, 0o644); err != nil {
			return err
		}
	}
	return nil
}
