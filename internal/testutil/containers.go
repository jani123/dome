//go:build e2e || integration

package testutil

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/jani123/dome/core/state"
	"github.com/moby/moby/client"
	tc "github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/wait"
)

func Root() string {
	_, source, _, _ := runtime.Caller(0)
	return filepath.Clean(filepath.Join(filepath.Dir(source), "../.."))
}

func BuildAgent(ctx context.Context, dir string) (string, error) {
	binary := filepath.Join(dir, "agent")
	cmd := exec.CommandContext(ctx, "go", "build", "-trimpath", "-o", binary, "./examples/basic-agent")
	cmd.Dir = Root()
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS=linux")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("build example: %w: %s", err, out)
	}
	return binary, nil
}

type AgentContainer struct {
	tc.Container
	Volume string
}

func StartAgent(ctx context.Context, binary string, networks []string) (*AgentContainer, error) {
	volume := "dome-test-" + state.NewID()
	request := tc.ContainerRequest{Image: "alpine:3.22", Entrypoint: []string{"/bin/sh"}, Cmd: []string{"-c", "mkdir -p /data && exec sleep infinity"}, Env: map[string]string{"DOME_ROOT": "/data"}, Files: []tc.ContainerFile{{HostFilePath: binary, ContainerFilePath: "/usr/local/bin/agent", FileMode: 0o755}}, Mounts: tc.ContainerMounts{tc.VolumeMount(volume, "/data")}, Networks: networks, WaitingFor: wait.ForExec([]string{"test", "-d", "/data"}).WithStartupTimeout(time.Minute)}
	c, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{ContainerRequest: request, Started: true})
	if err != nil {
		if c != nil {
			_ = c.Terminate(context.Background())
		}
		_ = removeVolume(context.Background(), volume)
		return nil, err
	}
	return &AgentContainer{c, volume}, nil
}

func removeVolume(ctx context.Context, volume string) error {
	provider, err := tc.NewDockerProvider()
	if err != nil {
		return err
	}
	defer provider.Close()
	_, err = provider.Client().VolumeRemove(ctx, volume, client.VolumeRemoveOptions{Force: true})
	return err
}

func (c *AgentContainer) Cleanup(ctx context.Context) error {
	if err := c.Terminate(ctx, tc.StopTimeout(time.Second)); err != nil {
		return err
	}
	return removeVolume(ctx, c.Volume)
}

func Exec(ctx context.Context, c tc.Container, args ...string) (int, string, error) {
	code, r, err := c.Exec(ctx, args, tcexec.Multiplexed())
	if err != nil {
		return code, "", err
	}
	b, err := io.ReadAll(r)
	return code, string(b), err
}

func (c *AgentContainer) Agent(ctx context.Context, args ...string) (int, string, error) {
	argv := append([]string{"agent", "--state", "/data/state.db", "--json"}, args...)
	return Exec(ctx, c.Container, argv...)
}
