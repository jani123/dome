package playbook_test

import (
	"context"
	"fmt"

	"github.com/jani123/dome/core/playbook"
)

func ExampleNew() {
	// Supply real injected operations in your application. Observe must report
	// actual progress; Recover must establish a safe state before another Apply.
	book := playbook.New("deploy", &playbook.Step{
		ID: "prepare", Repeatable: true,
		Observe: func(context.Context) (playbook.Status, error) { return playbook.Satisfied, nil },
		Apply:   func(context.Context) error { return nil },
		Verify:  func(context.Context) error { return nil },
	}, &playbook.Step{
		ID: "activate", Dependencies: []string{"prepare"}, Locks: []string{"app"},
		Observe: func(context.Context) (playbook.Status, error) { return playbook.Satisfied, nil },
		Apply:   func(context.Context) error { return nil },
		Verify:  func(context.Context) error { return nil },
		Recover: func(context.Context) error { return nil },
	})
	fmt.Println(book.Validate())
	// Output: <nil>
}
