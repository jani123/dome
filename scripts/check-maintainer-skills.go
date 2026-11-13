//go:build ignore

package main

import (
	"fmt"
	"os"

	"github.com/jani123/dome/internal/knowledge"
)

func main() {
	root, err := os.OpenRoot(".")
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot open repository")
		os.Exit(1)
	}
	defer root.Close()
	issues, err := knowledge.ValidateWorkflows(root.FS())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, issue := range issues {
		fmt.Fprintln(os.Stderr, issue)
	}
	if len(issues) > 0 {
		os.Exit(1)
	}
	fmt.Println("Maintainer skill metadata matches the pinned OpenSpec CLI")
}
