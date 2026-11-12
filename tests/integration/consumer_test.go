//go:build integration

package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/jani123/dome/internal/testutil"
	"github.com/stretchr/testify/require"
)

func TestExternalModuleConsumer(t *testing.T) {
	dir := t.TempDir()
	module := "module example.com/dome-consumer\n\ngo 1.26.0\n\nrequire github.com/jani123/dome v0.0.0\nreplace github.com/jani123/dome => " + strconv.Quote(testutil.Root()) + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte(module), 0o600))
	source := `package main
import("context";"fmt";"os";"github.com/jani123/dome/agent";"github.com/jani123/dome/config";"github.com/jani123/dome/core/resource")
type external struct{}
func(external)ID()resource.ID{return resource.ID{Module:"external",Name:"sample"}}
func(external)Validate()error{return nil}
func(external)Desired()resource.Snapshot{return resource.Snapshot{Exists:true,Value:"v1"}}
func(external)Observe(context.Context)(resource.Snapshot,error){return resource.Snapshot{},nil}
func(external)Apply(context.Context,resource.Action)error{return fmt.Errorf("plan must never call Apply")}
type module struct{}
func(module)Name()string{return "external"}
func(module)Restore(resource.ID)(resource.Resource,error){return external{},nil}
func main(){a,err:=agent.New(agent.WithModules(module{}),agent.WithDesiredState(func(context.Context,config.Document)([]resource.Resource,error){return []resource.Resource{external{}},nil}));if err==nil{err=a.Execute(context.Background())};if err!=nil{fmt.Fprintln(os.Stderr,err);os.Exit(1)}}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "main.go"), []byte(source), 0o600))
	binary := filepath.Join(dir, "consumer")
	cmd := exec.Command("go", "build", "-mod=mod", "-o", binary, ".")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	out, err = exec.Command(binary, "--state", filepath.Join(dir, "state.db"), "plan", "--json").CombinedOutput()
	require.NoError(t, err, string(out))
	require.Contains(t, string(out), `"external"`)
	require.Contains(t, string(out), `"create"`)
}
