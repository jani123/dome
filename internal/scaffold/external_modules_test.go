package scaffold

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/jani123/dome/core/fault"
	"github.com/stretchr/testify/require"
)

func TestExternalModuleRendering(t *testing.T) {
	files, err := RenderWithModules("example.com/consumer", "v0.1.0", []string{"example.com/one/widget@v1.2.3", "example.org/two/widget/v2@v2.1.0"})
	require.NoError(t, err)
	src := string(files["cmd/agent/main.go"])
	require.Contains(t, src, `external1 "example.com/one/widget"`)
	require.Contains(t, src, `external2 "example.org/two/widget/v2"`)
	require.Contains(t, src, "external1.Module{}")
	require.Contains(t, src, "external2.Module{}")
	_, err = parser.ParseFile(token.NewFileSet(), "main.go", src, parser.AllErrors)
	require.NoError(t, err)
	require.Contains(t, string(files["go.mod"]), "require example.com/one/widget v1.2.3")
	require.Contains(t, string(files["go.mod"]), "require example.org/two/widget/v2 v2.1.0")
	require.Contains(t, string(files["README.md"]), "required backend")
}

func TestExternalModuleValidation(t *testing.T) {
	for _, selections := range [][]string{
		{"example.com/widget"},
		{"example.com/widget@main"},
		{"example.com/widget@latest"},
		{"bad path@v1.0.0"},
		{"example.com/widget@v1.0"},
		{"example.com/widget@v1.0.0+meta"},
		{"example.com/widget@v2.0.0"},
		{"example.com/consumer@v1.0.0"},
		{"github.com/jani123/dome@v0.1.0"},
		{"github.com/jani123/dome/modules/file@v0.1.0"},
		{"example.com/widget@v1.0.0", "example.com/widget@v1.0.1"},
	} {
		_, err := RenderWithModules("example.com/consumer", "v0.1.0", selections)
		require.Error(t, err)
		require.Equal(t, string(fault.Validation), fault.Category(err))
	}
}

func TestPinnedDependencies(t *testing.T) {
	before := []byte("module example.com/consumer\nrequire (\ngithub.com/jani123/dome v0.1.0\nexample.com/widget v1.0.0\n)\n")
	require.NoError(t, checkPinnedDependencies(before, before))
	for _, after := range []string{
		"module example.com/consumer\nrequire github.com/jani123/dome v0.2.0\nrequire example.com/widget v1.0.0\n",
		"module example.com/consumer\nrequire github.com/jani123/dome v0.1.0\nrequire example.com/widget v1.1.0\n",
		"module example.com/consumer\nrequire github.com/jani123/dome v0.1.0\n",
	} {
		err := checkPinnedDependencies(before, []byte(after))
		require.Error(t, err)
		require.Equal(t, string(fault.Conflict), fault.Category(err))
	}
	require.Error(t, checkPinnedDependencies([]byte("invalid"), before))
	require.Error(t, checkPinnedDependencies(before, []byte("invalid")))
}
