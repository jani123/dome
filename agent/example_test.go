package agent_test

import (
	"context"
	"fmt"

	"github.com/jani123/dome/agent"
	"github.com/jani123/dome/config"
	"github.com/jani123/dome/core/resource"
	"github.com/jani123/dome/modules/file"
)

// Compose an agent without opening SQLite or applying host resources.
func ExampleNew() {
	type Settings struct {
		Content string `yaml:"content"`
	}
	type Secrets struct{}
	a, err := agent.New(
		agent.WithModules(file.Module{}),
		agent.WithDesiredState(func(ctx context.Context, doc config.Document) ([]resource.Resource, error) {
			settings, _, err := config.Parse[Settings, Secrets](ctx, doc, nil)
			if err != nil {
				return nil, err
			}
			return []resource.Resource{&file.File{Path: "/etc/my-app.conf", Content: []byte(settings.Content), Mode: 0o640}}, nil
		}),
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(a.Modules())
	// Output: [file]
}
