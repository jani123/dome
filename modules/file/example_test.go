package file_test

import (
	"fmt"

	"github.com/jani123/dome/modules/file"
)

func ExampleFile() {
	f := &file.File{Path: "/etc/my-app.conf", Content: []byte("hello\n"), Mode: 0o640}
	fmt.Println(f.Validate())
	fmt.Println(f.ID())
	fmt.Println(f.Desired().Exists)
	// Output:
	// <nil>
	// file:/etc/my-app.conf
	// true
}
