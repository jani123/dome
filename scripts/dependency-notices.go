//go:build ignore

// Generate license/notice material for actual build dependencies, not every
// module in the graph. Run separately for every published binary package.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type (
	module struct {
		Path, Version, Dir string
		Main               bool
	}
	pkg struct{ Module *module }
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != 3 {
		return fmt.Errorf("usage: go run scripts/dependency-notices.go <package> <output>")
	}
	cmd := exec.Command("go", "list", "-deps", "-json", os.Args[1])
	if target := os.Getenv("DOME_NOTICE_GOARCH"); target != "" {
		cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH="+target, "CGO_ENABLED=0")
	}
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(out))
	modules := map[string]*module{}
	for {
		var p pkg
		err := decoder.Decode(&p)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if p.Module != nil && !p.Module.Main {
			modules[p.Module.Path] = p.Module
		}
	}
	names := make([]string, 0, len(modules))
	for name := range modules {
		names = append(names, name)
	}
	sort.Strings(names)
	var report bytes.Buffer
	fmt.Fprintln(&report, "Third-party license and attribution material\nGenerated for", os.Args[1])
	for _, name := range names {
		m := modules[name]
		entries, err := os.ReadDir(m.Dir)
		if err != nil {
			return err
		}
		found := false
		fmt.Fprintf(&report, "\n===== %s %s =====\n", m.Path, m.Version)
		for _, entry := range entries {
			upper := strings.ToUpper(entry.Name())
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if entry.IsDir() || (!strings.HasPrefix(upper, "LICENSE") && !strings.HasPrefix(upper, "COPYING") && !strings.HasPrefix(upper, "NOTICE")) || ext == ".go" {
				continue
			}
			body, err := os.ReadFile(filepath.Join(m.Dir, entry.Name()))
			if err != nil {
				return err
			}
			fmt.Fprintf(&report, "\n--- %s ---\n%s\n", entry.Name(), body)
			if strings.HasPrefix(upper, "LICENSE") || strings.HasPrefix(upper, "COPYING") {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("missing license material for %s; review before distribution", name)
		}
	}
	// The compiler/runtime are distributed under the Go license too.
	gorootCmd := exec.Command("go", "env", "GOROOT")
	goroot, err := gorootCmd.Output()
	if err != nil {
		return err
	}
	license, err := os.ReadFile(filepath.Join(strings.TrimSpace(string(goroot)), "LICENSE"))
	if err != nil {
		return err
	}
	fmt.Fprintf(&report, "\n===== Go toolchain/runtime =====\n%s\n", license)
	return os.WriteFile(os.Args[2], report.Bytes(), 0o644)
}
