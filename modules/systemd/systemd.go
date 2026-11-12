// Package systemd manages service enabled/running state through an injectable runner.
package systemd

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/jani123/dome/core/fault"
	"github.com/jani123/dome/core/resource"
)

type (
	// Runner isolates systemctl execution for tests.
	Runner interface {
		Run(context.Context, string, ...string) ([]byte, error)
	}
	// OSRunner executes commands on the target host.
	OSRunner struct{}
)

// Run executes a host command with context cancellation.
func (OSRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).Output()
}

// Service declares enabled and running state of a loaded .service unit.
type Service struct {
	Name    string
	Enabled bool
	Running bool
	Runner  Runner
}

var serviceName = regexp.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9_.@:-]*\.service$`)

// ID uses the service unit name as its stable identity.
func (s *Service) ID() resource.ID { return resource.ID{Module: "systemd", Name: s.Name} }

// Validate restricts management to valid .service unit names.
func (s *Service) Validate() (err error) {
	defer func() { err = fault.Wrap(fault.Validation, err, "module", "systemd", "resource_id", s.ID().String()) }()
	if !serviceName.MatchString(s.Name) {
		return fmt.Errorf("service name must be a valid .service unit")
	}
	return nil
}

func (s *Service) runner() Runner {
	if s.Runner != nil {
		return s.Runner
	}
	return OSRunner{}
}

func snapshot(enabled, running bool) resource.Snapshot {
	return resource.Snapshot{Exists: true, Value: fmt.Sprintf("enabled=%t;running=%t", enabled, running)}
}

// Desired describes enabled and running state without secret data.
func (s *Service) Desired() resource.Snapshot { return snapshot(s.Enabled, s.Running) }

// Observe inspects loaded service state and rejects unsupported unit states.
func (s *Service) Observe(ctx context.Context) (_ resource.Snapshot, err error) {
	defer func() { err = fault.Wrap(fault.Observation, err, "module", "systemd", "resource_id", s.ID().String()) }()
	if err = s.Validate(); err != nil {
		return resource.Snapshot{}, err
	}
	b, err := s.runner().Run(ctx, "systemctl", "show", "--property=LoadState,ActiveState,UnitFileState", "--", s.Name)
	if err != nil {
		return resource.Snapshot{}, err
	}
	props := map[string]string{}
	for _, line := range strings.Split(string(b), "\n") {
		k, v, ok := strings.Cut(line, "=")
		if ok {
			props[k] = v
		}
	}
	if props["LoadState"] != "loaded" {
		return resource.Snapshot{}, fmt.Errorf("service is not loaded")
	}
	active := props["ActiveState"]
	if active != "active" && active != "inactive" && active != "failed" {
		return resource.Snapshot{}, fmt.Errorf("service is in a transitional or unsupported state")
	}
	enabled := props["UnitFileState"]
	if enabled != "enabled" && enabled != "disabled" {
		return resource.Snapshot{}, fmt.Errorf("unit enabled state is unsupported")
	}
	return snapshot(enabled == "enabled", active == "active"), nil
}

// Apply reconciles service state; deletion stops and disables without uninstalling.
func (s *Service) Apply(ctx context.Context, action resource.Action) (err error) {
	defer func() { err = fault.Wrap(fault.Apply, err, "module", "systemd", "resource_id", s.ID().String()) }()
	if err = s.Validate(); err != nil {
		return err
	}
	if action == resource.Noop {
		return nil
	}
	if action != resource.Create && action != resource.Update && action != resource.Delete {
		return fault.New(fault.Validation, "unsupported systemd action")
	}
	enabled, running := s.Enabled, s.Running
	if action == resource.Delete {
		enabled = false
		running = false
	}
	commands := []string{}
	if !running {
		commands = append(commands, "stop")
	}
	if enabled {
		commands = append(commands, "enable")
	} else {
		commands = append(commands, "disable")
	}
	if running {
		commands = append(commands, "start")
	}
	for _, command := range commands {
		if _, err := s.runner().Run(ctx, "systemctl", command, "--", s.Name); err != nil {
			return err
		}
	}
	return nil
}

// A pruned service is stopped and disabled, never uninstalled.
type Module struct{ Runner Runner }

// Name identifies the explicitly registered systemd module.
func (Module) Name() string { return "systemd" }

// Restore validates an owned service identity and constructs a stop/disable handle.
func (m Module) Restore(id resource.ID) (resource.Resource, error) {
	if id.Module != "systemd" {
		return nil, fmt.Errorf("wrong module")
	}
	s := &Service{Name: id.Name, Runner: m.Runner}
	return s, s.Validate()
}
