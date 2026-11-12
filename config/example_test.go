package config_test

import (
	"context"
	"fmt"

	"github.com/jani123/dome/config"
)

func ExampleParse() {
	type Settings struct {
		Content string `yaml:"content"`
	}
	type Secrets struct{}
	settings, _, err := config.Parse[Settings, Secrets](context.Background(), config.Document{Configuration: []byte("content: hello\n")}, nil)
	fmt.Println(settings.Content, err)
	// Output: hello <nil>
}
