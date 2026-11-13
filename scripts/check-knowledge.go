//go:build ignore

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jani123/dome/internal/knowledge"
)

func main() {
	policy := flag.Bool("policy", false, "also enforce Dome provenance, index and local-link policy")
	flag.Parse()
	bundle := "knowledge"
	if flag.NArg() > 1 {
		fmt.Fprintln(os.Stderr, "usage: check-knowledge [-policy] [bundle]")
		os.Exit(2)
	}
	if flag.NArg() == 1 {
		bundle = flag.Arg(0)
	}
	root, err := os.OpenRoot(".")
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot open repository")
		os.Exit(1)
	}
	defer root.Close()
	issues, err := knowledge.Validate(root.FS(), knowledge.Options{Bundle: bundle, Policy: *policy})
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
	fmt.Println("OKF v0.2 validation passed; repository policy:", *policy)
}
