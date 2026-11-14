package scaffold

import (
	"embed"
	"encoding/json"
	"go/format"
	"go/token"
	"io/fs"
	"regexp"
	"strings"
	"time"

	"github.com/jani123/dome/core/fault"
	"golang.org/x/mod/module"
)

//go:embed all:moduletemplates
var moduleTemplates embed.FS

// File includes executable modes required by generated developer tooling.
type File struct {
	Data []byte
	Mode fs.FileMode
}

// RenderModule renders an external resource library without touching the host.
func RenderModule(modulePath, name, version string) (map[string]File, error) {
	// Keep supported release validation identical to consumer generation.
	if _, err := Render(modulePath, version); err != nil {
		return nil, err
	}
	_, major, ok := module.SplitPathVersion(modulePath)
	if !ok || major != "" {
		return nil, fault.New(fault.Validation, "module starter supports v0/v1 import paths only")
	}
	if !regexp.MustCompile(`^[a-z][a-z0-9_]*$`).MatchString(name) || token.Lookup(name).IsKeyword() || name == "main" || name == "init" {
		return nil, fault.New(fault.Validation, "name must be an importable lowercase Go package identifier")
	}
	generatedAt := time.Now().UTC().Format(time.RFC3339)
	files := map[string]File{}
	err := fs.WalkDir(moduleTemplates, "moduletemplates", func(path string, e fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if e.IsDir() {
			return nil
		}
		b, err := moduleTemplates.ReadFile(path)
		if err != nil {
			return err
		}
		text := strings.NewReplacer("example.com/__MODULE__", modulePath, "__MODULE__", modulePath, "__NAME__", name, "__VERSION__", version, "__GENERATED_AT__", generatedAt).Replace(string(b))
		out := strings.TrimSuffix(strings.TrimPrefix(path, "moduletemplates/"), ".tmpl")
		out = strings.ReplaceAll(out, "_agents/", ".agents/")
		out = strings.ReplaceAll(out, "_github/", ".github/")
		out = strings.ReplaceAll(out, "_devcontainer/", ".devcontainer/")
		if out == "gitignore" {
			out = ".gitignore"
		}
		if out == "openspec-version" {
			out = ".openspec-version"
		}
		if out == "hugo-version" {
			out = ".hugo-version"
		}
		b = []byte(text)
		if strings.HasSuffix(out, ".go") {
			b, err = format.Source(b)
			if err != nil {
				return err
			}
		}
		mode := fs.FileMode(0o644)
		if strings.HasSuffix(out, ".sh") {
			mode = 0o755
		}
		files[out] = File{b, mode}
		return nil
	})
	if err != nil {
		return nil, fault.Wrap(fault.Configuration, err)
	}
	b, err := json.MarshalIndent(map[string]any{"schema": 1, "module": modulePath, "name": name, "dome_version": version, "skill_version": version, "template_version": version}, "", "  ")
	if err != nil {
		return nil, err
	}
	files["dome.json"] = File{append(b, '\n'), 0o644}
	return files, nil
}
