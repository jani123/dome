// Package scaffold renders a version-matched external Dome consumer.
package scaffold

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"go/format"
	"io/fs"
	"strings"
	"text/template"

	"github.com/jani123/dome/core/fault"
	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
)

//go:embed all:templates
var templates embed.FS

// Render is pure: callers publish the returned files only after validation.
func Render(modulePath, version string) (map[string][]byte, error) {
	if err := module.CheckPath(modulePath); err != nil {
		return nil, fault.Wrap(fault.Validation, err)
	}
	if modulePath == "github.com/jani123/dome" || strings.HasPrefix(modulePath, "github.com/jani123/dome/") {
		return nil, fault.New(fault.Validation, "choose a consumer module outside the Dome module path")
	}
	if !semver.IsValid(version) || semver.Canonical(version) != version || semver.Major(version) != "v0" || semver.Build(version) != "" {
		return nil, fault.New(fault.Validation, "initializer requires an exact v0 release or resolved Go pseudo-version; use a versioned initializer")
	}
	files := map[string][]byte{}
	guideVersion := version
	if module.IsPseudoVersion(version) {
		guideVersion = "next"
	}
	data := struct{ Module, Version, GuideVersion string }{modulePath, version, guideVersion}
	err := fs.WalkDir(templates, "templates", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		body, err := templates.ReadFile(path)
		if err != nil {
			return err
		}
		t, err := template.New(path).Option("missingkey=error").Parse(string(body))
		if err != nil {
			return err
		}
		var out bytes.Buffer
		if err := t.Execute(&out, data); err != nil {
			return err
		}
		name := strings.TrimSuffix(strings.TrimPrefix(path, "templates/"), ".tmpl")
		// Embedded template names avoid hidden-file exclusions in go:embed.
		name = strings.ReplaceAll(name, "_agents/", ".agents/")
		if name == "gitignore" {
			name = ".gitignore"
		}
		content := out.Bytes()
		if strings.HasSuffix(name, ".go") {
			content, err = format.Source(content)
			if err != nil {
				return err
			}
		}
		files[name] = content
		return nil
	})
	if err != nil {
		return nil, fault.Wrap(fault.Configuration, err)
	}
	manifest, err := json.MarshalIndent(struct {
		Schema   int    `json:"schema"`
		Module   string `json:"module"`
		Version  string `json:"dome_version"`
		Skill    string `json:"skill_version"`
		Template string `json:"template_version"`
	}{1, modulePath, version, version, version}, "", "  ")
	if err != nil {
		return nil, fault.Wrap(fault.Configuration, err)
	}
	files["dome.json"] = append(manifest, '\n')
	return files, nil
}

// CheckVersion detects dependency/asset drift before authoring or updates.
func CheckVersion(dependency string, manifest []byte) error {
	var m struct {
		Version  string `json:"dome_version"`
		Skill    string `json:"skill_version"`
		Template string `json:"template_version"`
	}
	if err := json.Unmarshal(manifest, &m); err != nil {
		return fault.Wrap(fault.Validation, err)
	}
	if dependency != m.Version || dependency != m.Skill || dependency != m.Template {
		return fault.New(fault.Validation, fmt.Sprintf("Dome version mismatch: regenerate a temporary project using the initializer at %s, review its skill/templates, and update dependency and assets together", dependency))
	}
	return nil
}
