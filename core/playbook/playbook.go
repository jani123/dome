// Package playbook defines explicit, typed workflow graphs.
package playbook

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sort"
	"time"
)

// Status describes observation results used to decide safe execution or recovery.
type Status string

const (
	Satisfied  Status = "satisfied"
	NeedsApply Status = "needs-apply"
	InProgress Status = "in-progress"
	Failed     Status = "failed"
	Unknown    Status = "unknown"
)

type (
	// Ref names a registered playbook.
	Ref string
	// RetryPolicy bounds attempts per invocation and delays between retries.
	RetryPolicy struct {
		Attempts int
		Delay    time.Duration
	}
	// Observe and Verify must be read-only. A non-repeatable interrupted step must
	// supply Recover to establish a safe state before Apply can run again.
	Step struct {
		ID           string
		Dependencies []string
		Locks        []string
		Observe      func(context.Context) (Status, error)
		Apply        func(context.Context) error
		Verify       func(context.Context) error
		Recover      func(context.Context) error
		Repeatable   bool
		Timeout      time.Duration
		Retry        RetryPolicy
		// OnFailure is an explicit recovery branch. It cannot hide the original failure.
		OnFailure *Step
	}
)

// Playbook is a versioned explicit DAG; bump Version when callback behavior changes.
type Playbook struct {
	Name    string
	Version string
	Steps   []*Step
}

// New constructs a workflow at version 1; validate it before execution.
func New(name string, steps ...*Step) *Playbook {
	return &Playbook{Name: name, Version: "1", Steps: steps}
}

// Validate checks graph structure, callback requirements, and safe recovery contracts.
func (p *Playbook) Validate() error {
	if p == nil || p.Name == "" {
		return fmt.Errorf("playbook name is required")
	}
	nodes := map[string]*Step{}
	for _, s := range p.Steps {
		if s == nil || s.ID == "" || s.Observe == nil || s.Apply == nil || s.Verify == nil {
			return fmt.Errorf("each step requires ID, Observe, Apply, and Verify")
		}
		if _, ok := nodes[s.ID]; ok {
			return fmt.Errorf("duplicate step %s", s.ID)
		}
		if s.Timeout < 0 || s.Retry.Attempts < 0 || s.Retry.Delay < 0 {
			return fmt.Errorf("invalid timeout or retry policy")
		}
		if !s.Repeatable && s.Recover == nil {
			return fmt.Errorf("step %s must be repeatable or provide recovery", s.ID)
		}
		nodes[s.ID] = s
	}
	colors := map[string]int{}
	var visit func(string) error
	visit = func(id string) error {
		if colors[id] == 1 {
			return fmt.Errorf("cycle at step %s", id)
		}
		if colors[id] == 2 {
			return nil
		}
		s, ok := nodes[id]
		if !ok {
			return fmt.Errorf("missing dependency %s", id)
		}
		colors[id] = 1
		for _, d := range s.Dependencies {
			if err := visit(d); err != nil {
				return err
			}
		}
		colors[id] = 2
		return nil
	}
	for _, s := range p.Steps {
		if err := visit(s.ID); err != nil {
			return err
		}
		if s.OnFailure != nil {
			r := s.OnFailure
			if r.OnFailure != nil || len(r.Dependencies) != 0 {
				return fmt.Errorf("recovery branches cannot have dependencies or nested branches")
			}
			if err := New("recovery", r).Validate(); err != nil {
				return err
			}
			if _, ok := nodes[s.ID+"/recovery"]; ok {
				return fmt.Errorf("recovery node identity collides")
			}
		}
	}
	return nil
}

// Fingerprint rejects resuming a changed graph. Bump Version when callbacks change.
func (p *Playbook) Fingerprint() string {
	parts := []string{p.Name, p.Version}
	for _, s := range p.Steps {
		deps := append([]string(nil), s.Dependencies...)
		sort.Strings(deps)
		locks := append([]string(nil), s.Locks...)
		sort.Strings(locks)
		recovery := ""
		if s.OnFailure != nil {
			recovery = New("recovery", s.OnFailure).Fingerprint()
		}
		parts = append(parts, fmt.Sprintf("%s|%v|%v|%t|%d|%v|%s", s.ID, deps, locks, s.Repeatable, s.Timeout, s.Retry, recovery))
	}
	sort.Strings(parts[2:])
	return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprint(parts))))
}
