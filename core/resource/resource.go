// Package resource defines the public, module-independent resource contract.
package resource

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
)

// ID is a stable non-secret module/name resource identity.
type ID struct {
	Module string `json:"module"`
	Name   string `json:"name"`
}

// String formats a stable identity; module and name must contain no secrets.
func (id ID) String() string { return id.Module + ":" + id.Name }

// Validate checks required identity fields and module syntax.
func (id ID) Validate() error {
	if id.Module == "" || id.Name == "" || strings.ContainsAny(id.Module, ":/ \t\n") {
		return fmt.Errorf("resource module and name are required")
	}
	return nil
}

// Snapshot contains comparable, non-secret state (for example content digests).
// Do not put plaintext secrets in snapshots: they are persisted and shown in plans.
type Snapshot struct {
	Exists bool   `json:"exists"`
	Value  string `json:"value,omitempty"`
}

// Fingerprint hashes the fixed snapshot representation for durable transition identities.
func (s Snapshot) Fingerprint() string {
	b, _ := json.Marshal(s)
	return fmt.Sprintf("%x", sha256.Sum256(b))
}

// Action describes a change between observed and desired snapshots.
type Action string

const (
	Create Action = "create"
	Update Action = "update"
	Delete Action = "delete"
	Noop   Action = "no-op"
)

// Difference compares observed and desired snapshots without consulting persisted ownership.
func Difference(actual, desired Snapshot) Action {
	if actual == desired {
		return Noop
	}
	if !desired.Exists {
		return Delete
	}
	if !actual.Exists {
		return Create
	}
	return Update
}

// Resource implementations must make Apply repeatable and Observe truthful.
// Validate must reject unsafe inputs before any host mutation takes place.
type Resource interface {
	ID() ID
	Validate() error
	Desired() Snapshot
	Observe(context.Context) (Snapshot, error)
	Apply(context.Context, Action) error
}

// Transition chooses a playbook for a change. If none is chosen, SafeDefault
// must be true before the reconciler can call the resource's Apply method.
type Transition interface {
	Transition(Snapshot) (playbook string, safeDefault bool, err error)
}

// Module restores deletion-only handles for resources previously owned by it.
// Restore must validate the identity and must never require old secret contents.
type Module interface {
	Name() string
	Restore(ID) (Resource, error)
}
