//go:build ignore

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jani123/dome/internal/review"
)

func main() {
	aggregate := flag.Bool("aggregate", false, "validate consolidated report")
	reviewID := flag.String("review-id", "", "expected review id")
	inputID := flag.String("input-id", "", "expected input id")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: check-review-report [--aggregate] [--review-id ID] [--input-id ID] report.json")
		os.Exit(2)
	}
	if err := run(flag.Arg(0), *aggregate, *reviewID, *inputID); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Review report validated")
}

func run(path string, aggregate bool, reviewID, inputID string) error {
	name := "report"
	if aggregate {
		name = "aggregate"
	}
	schema, err := os.ReadFile(".agents/skills/dome-review/references/" + name + ".schema.json")
	if err != nil {
		return err
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	r, err := review.Validate(schema, body, reviewID, inputID)
	if err != nil {
		return err
	}
	if !aggregate {
		return nil
	}
	root, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer root.Close()
	childSchema, err := os.ReadFile(".agents/skills/dome-review/references/report.schema.json")
	if err != nil {
		return err
	}
	reports := map[string]review.Report{}
	for _, role := range r.Roles {
		if role.Reference == nil {
			continue
		}
		if !filepath.IsLocal(*role.Reference) {
			return fmt.Errorf("report reference must remain local")
		}
		b, err := root.ReadFile(*role.Reference)
		if err != nil {
			return err
		}
		child, err := review.Validate(childSchema, b, r.ReviewID, r.InputID)
		if err != nil {
			return err
		}
		if child.Role != role.Role || role.ReportID == nil || child.ReportID != *role.ReportID || child.Outcome != role.Outcome {
			return fmt.Errorf("role report identity/outcome mismatch")
		}
		if _, exists := reports[child.ReportID]; exists {
			return fmt.Errorf("duplicate report id")
		}
		reports[child.ReportID] = child
	}
	for _, f := range r.Findings {
		if len(f.Origins) == 0 {
			return fmt.Errorf("aggregate finding lacks provenance")
		}
		for _, o := range f.Origins {
			child, ok := reports[o.ReportID]
			if !ok {
				return fmt.Errorf("missing origin report")
			}
			found := false
			for _, cf := range child.Findings {
				if cf.ID == o.FindingID {
					found = true
				}
			}
			if !found {
				return fmt.Errorf("missing origin finding")
			}
		}
	}
	return nil
}
