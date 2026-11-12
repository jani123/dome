//go:build ignore

// Check generated internal links against a project-site URL prefix.
package main

import (
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != 3 {
		return fmt.Errorf("usage: go run scripts/check-docs.go <output-dir> <url-prefix>")
	}
	root, prefix := os.Args[1], os.Args[2]
	re := regexp.MustCompile(`(?:href|src)="([^"]+)"`)
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".html") {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, match := range re.FindAllSubmatch(body, -1) {
			u, err := url.Parse(string(match[1]))
			if err != nil {
				return err
			}
			if u.IsAbs() || u.Host != "" || u.Path == "" {
				continue
			}
			target := filepath.Join(filepath.Dir(path), u.Path)
			if strings.HasPrefix(u.Path, "/") {
				if !strings.HasPrefix(u.Path, prefix) {
					return fmt.Errorf("%s: link outside site prefix: %s", path, u.Path)
				}
				target = filepath.Join(root, strings.TrimPrefix(u.Path, prefix))
			}
			if strings.HasSuffix(u.Path, "/") {
				target = filepath.Join(target, "index.html")
			}
			if _, err := os.Stat(target); err != nil {
				return fmt.Errorf("%s: broken link %s: %w", path, u.Path, err)
			}
		}
		return nil
	})
}
