// basic-agent is a standalone consumer of the public framework packages.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/jani123/dome/agent"
	"github.com/jani123/dome/cli"
	"github.com/jani123/dome/config"
	"github.com/jani123/dome/core/playbook"
	"github.com/jani123/dome/core/resource"
	"github.com/jani123/dome/modules/file"
	"github.com/jani123/dome/modules/systemd"
)

type (
	settings struct {
		Root           string `yaml:"root"`
		Content        string `yaml:"content"`
		Managed        *bool  `yaml:"managed"`
		Version        string `yaml:"version"`
		FailTransition bool   `yaml:"fail_transition"`
	}
	secrets struct {
		Token string `yaml:"token"`
	}
	versionResource struct {
		file.File
		Fail bool
	}
)

func (v *versionResource) ID() resource.ID { return resource.ID{Module: "demo-version", Name: v.Path} }

func (v *versionResource) Transition(resource.Snapshot) (string, bool, error) {
	if v.Absent {
		return "", true, nil
	}
	return "set-version", false, nil
}

type versionModule struct{}

func (versionModule) Name() string { return "demo-version" }
func (versionModule) Restore(id resource.ID) (resource.Resource, error) {
	v := &versionResource{File: file.File{Path: id.Name, Absent: true}}
	return v, v.Validate()
}

func root() string {
	if p := os.Getenv("DOME_ROOT"); p != "" {
		return p
	}
	return "/tmp/dome-example"
}

func markerStep(id, name string, delay time.Duration) *playbook.Step {
	path := func() string { return filepath.Join(root(), name) }
	observe := func(ctx context.Context) (playbook.Status, error) {
		_, err := os.Stat(path())
		if errors.Is(err, os.ErrNotExist) {
			return playbook.NeedsApply, nil
		}
		if err != nil {
			return playbook.Unknown, err
		}
		return playbook.Satisfied, nil
	}
	return &playbook.Step{ID: id, Repeatable: true, Timeout: 10 * time.Second, Observe: observe, Apply: func(ctx context.Context) error {
		if delay > 0 {
			timer := time.NewTimer(delay)
			defer timer.Stop()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
			}
		}
		if err := os.WriteFile(path(), []byte(id+"\n"), 0o600); err != nil {
			return err
		}
		if id == "external-write" && os.Getenv("DOME_CRASH_AFTER_WRITE") == "1" {
			os.Exit(77)
		}
		return nil
	}, Verify: func(ctx context.Context) error {
		s, err := observe(ctx)
		if err != nil {
			return err
		}
		if s != playbook.Satisfied {
			return fmt.Errorf("marker verification failed")
		}
		return nil
	}}
}

func main() {
	var target *versionResource
	transition := playbook.New("set-version", &playbook.Step{ID: "write-version", Repeatable: true, Observe: func(ctx context.Context) (playbook.Status, error) {
		if target == nil {
			return playbook.Unknown, fmt.Errorf("version is not configured")
		}
		s, err := target.Observe(ctx)
		if err != nil {
			return playbook.Unknown, err
		}
		if s == target.Desired() {
			return playbook.Satisfied, nil
		}
		return playbook.NeedsApply, nil
	}, Apply: func(ctx context.Context) error {
		if target.Fail {
			return fmt.Errorf("demonstration transition failure")
		}
		if err := target.Apply(ctx, resource.Update); err != nil {
			return err
		}
		if os.Getenv("DOME_CRASH_AFTER_VERSION_WRITE") == "1" {
			os.Exit(78)
		}
		return nil
	}, Verify: func(ctx context.Context) error {
		s, err := target.Observe(ctx)
		if err != nil {
			return err
		}
		if s != target.Desired() {
			return fmt.Errorf("version verification failed")
		}
		return nil
	}})
	// Finalization is a separate workflow effect: reaching the target version
	// alone is not proof that the transition completed.
	finalized := func(ctx context.Context) (playbook.Status, error) {
		if target == nil {
			return playbook.Unknown, fmt.Errorf("version is not configured")
		}
		b, err := os.ReadFile(target.Path + ".finalized")
		if errors.Is(err, os.ErrNotExist) {
			return playbook.NeedsApply, nil
		}
		if err != nil {
			return playbook.Unknown, err
		}
		if string(b) == string(target.Content) {
			return playbook.Satisfied, nil
		}
		return playbook.NeedsApply, nil
	}
	transition.Steps = append(transition.Steps, &playbook.Step{ID: "finalize-version", Dependencies: []string{"write-version"}, Repeatable: true, Observe: finalized, Apply: func(ctx context.Context) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		return os.WriteFile(target.Path+".finalized", target.Content, 0o600)
	}, Verify: func(ctx context.Context) error {
		s, err := finalized(ctx)
		if err != nil {
			return err
		}
		if s != playbook.Satisfied {
			return fmt.Errorf("version transition is not finalized")
		}
		return nil
	}})
	left := markerStep("left", "left.marker", 300*time.Millisecond)
	right := markerStep("right", "right.marker", 300*time.Millisecond)
	join := markerStep("join", "join.marker", 0)
	join.Dependencies = []string{"left", "right"}
	parallel := playbook.New("parallel-demo", left, right, join)
	crash := playbook.New("crash-demo", markerStep("external-write", "external.marker", 0))
	a, err := agent.New(agent.WithModules(file.Module{}, systemd.Module{}, versionModule{}), agent.WithPlaybooks(transition, parallel, crash), agent.WithDesiredState(func(ctx context.Context, d config.Document) ([]resource.Resource, error) {
		c, _, err := config.Parse[settings, secrets](ctx, d, nil)
		if err != nil {
			return nil, err
		}
		if c.Root == "" {
			c.Root = root()
		}
		if !filepath.IsAbs(c.Root) || filepath.Clean(c.Root) != c.Root {
			return nil, fmt.Errorf("root must be a clean absolute directory")
		}
		info, err := os.Stat(c.Root)
		if err != nil || !info.IsDir() {
			return nil, fmt.Errorf("root must be an existing directory")
		}
		resources := []resource.Resource{}
		if c.Managed == nil || *c.Managed {
			resources = append(resources, &file.File{Path: filepath.Join(c.Root, "managed.txt"), Content: []byte(c.Content), Mode: 0o640})
		}
		target = nil
		if c.Version != "" {
			if strings.ContainsAny(c.Version, "\n\r") {
				return nil, fmt.Errorf("version must be a single line")
			}
			target = &versionResource{File: file.File{Path: filepath.Join(c.Root, "version.txt"), Content: []byte(c.Version), Mode: 0o600}, Fail: c.FailTransition}
			resources = append(resources, target)
		}
		return resources, nil
	}))
	if err == nil {
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()
		err = a.Execute(ctx)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, agent.ErrorText(err))
		os.Exit(cli.ExitCode(err))
	}
}
