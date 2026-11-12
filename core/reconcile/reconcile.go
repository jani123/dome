// Package reconcile implements Observe → Decide → Apply → Verify → Persist.
package reconcile

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/jani123/dome/core/fault"
	"github.com/jani123/dome/core/playbook"
	"github.com/jani123/dome/core/registry"
	"github.com/jani123/dome/core/resource"
	"github.com/jani123/dome/core/state"
)

type Repository interface {
	Managed(context.Context) ([]state.Managed, error)
	SaveManaged(context.Context, state.Managed) error
	Forget(context.Context, resource.ID) error
	SetRevision(context.Context, string) error
	BeginOperation(context.Context, string, string) (string, error)
	EndOperation(context.Context, string, string, string) error
	MetadataByPrefix(context.Context, string) (map[string]string, error)
	SetMetadata(context.Context, string, string) error
}
type (
	PlaybookRunner interface {
		Run(context.Context, *playbook.Playbook, string) (string, error)
	}
	Change struct {
		ID       resource.ID       `json:"id"`
		Action   resource.Action   `json:"action"`
		Before   resource.Snapshot `json:"before"`
		After    resource.Snapshot `json:"after"`
		Prune    bool              `json:"prune,omitempty"`
		Playbook string            `json:"playbook,omitempty"`
		Resume   string            `json:"execution_id,omitempty"`
		target   resource.Resource
	}
	Plan struct {
		Revision string   `json:"revision"`
		Changes  []Change `json:"changes"`
	}
	Reconciler struct {
		Store     Repository
		Registry  *registry.Registry
		Runner    PlaybookRunner
		Playbooks map[string]*playbook.Playbook
	}
)

// Plan validates and observes the entire desired state before allowing mutations.
func (r *Reconciler) Plan(ctx context.Context, desired []resource.Resource, revision string) (*Plan, error) {
	seen := map[resource.ID]bool{}
	for _, v := range desired {
		if v == nil {
			return nil, fault.New(fault.Validation, "nil desired resource")
		}
		id := v.ID()
		if err := id.Validate(); err != nil {
			return nil, fault.Wrap(fault.Validation, err)
		}
		if !r.Registry.Has(id.Module) {
			return nil, fault.New(fault.Validation, "resource module is not registered", "module", id.Module)
		}
		if seen[id] {
			return nil, fault.New(fault.Validation, "duplicate resource identity", "resource_id", id.String())
		}
		seen[id] = true
		if err := v.Validate(); err != nil {
			return nil, fault.Wrap(fault.Validation, err, "module", id.Module, "resource_id", id.String())
		}
	}
	pending, err := r.pendingTransitions(ctx, desired)
	if err != nil {
		return nil, err
	}
	managed, err := r.Store.Managed(ctx)
	if err != nil {
		return nil, fault.Wrap(fault.Persistence, err)
	}
	owned := map[resource.ID]bool{}
	for _, v := range managed {
		owned[v.ID] = true
	}
	p := &Plan{Revision: revision, Changes: []Change{}}
	add := func(v resource.Resource, prune bool) error {
		actual, err := v.Observe(ctx)
		if err != nil {
			return fault.Wrap(fault.Observation, err, "module", v.ID().Module, "resource_id", v.ID().String())
		}
		want := v.Desired()
		action := resource.Difference(actual, want)
		if prune && actual != want {
			action = resource.Delete
		}
		if action == resource.Delete && !owned[v.ID()] {
			return fault.New(fault.Validation, "refusing to delete an unowned resource", "resource_id", v.ID().String())
		}
		change := Change{ID: v.ID(), Action: action, Before: actual, After: want, Prune: prune, target: v}
		if action != resource.Noop {
			if transition, ok := v.(resource.Transition); ok {
				name, safe, err := transition.Transition(actual)
				if err != nil {
					return fault.Wrap(fault.Validation, err, "resource_id", v.ID().String())
				}
				if name != "" {
					book, ok := r.Playbooks[name]
					if !ok || r.Runner == nil {
						return fault.New(fault.Validation, "transition playbook is unavailable", "resource_id", v.ID().String())
					}
					if err = book.Validate(); err != nil {
						return fault.Wrap(fault.Validation, err)
					}
					change.Playbook = name
				} else if !safe {
					return fault.New(fault.Validation, "unsafe transition requires a playbook", "resource_id", v.ID().String())
				}
			}
		}
		if record, ok := pending[v.ID()]; ok {
			if change.Playbook != "" && change.Playbook != record.Playbook {
				return fault.New(fault.Recovery, "unfinished transition uses a different playbook", "resource_id", v.ID().String(), "operation_id", record.Execution)
			}
			book, ok := r.Playbooks[record.Playbook]
			if !ok || r.Runner == nil {
				return fault.New(fault.Recovery, "unfinished transition playbook is unavailable", "resource_id", v.ID().String(), "operation_id", record.Execution)
			}
			if err := book.Validate(); err != nil {
				return fault.Wrap(fault.Validation, err, "module", v.ID().Module, "resource_id", v.ID().String())
			}
			change.Playbook = record.Playbook
			change.Resume = record.Execution
		}
		p.Changes = append(p.Changes, change)
		return nil
	}
	for _, v := range desired {
		if err = add(v, false); err != nil {
			return nil, err
		}
	}
	for _, v := range managed {
		if seen[v.ID] {
			continue
		}
		handle, err := r.Registry.Restore(v.ID)
		if err != nil {
			return nil, err
		}
		if err = add(handle, true); err != nil {
			return nil, err
		}
	}
	// A crash after Forget but before clearing the recovery intent must still
	// resume a pruning workflow whose ownership row was already removed.
	for id, record := range pending {
		if record.Prune && !seen[id] && !owned[id] {
			handle, err := r.Registry.Restore(id)
			if err != nil {
				return nil, err
			}
			if err = add(handle, true); err != nil {
				return nil, err
			}
		}
	}
	sort.SliceStable(p.Changes, func(i, j int) bool {
		if p.Changes[i].Prune != p.Changes[j].Prune {
			return !p.Changes[i].Prune
		}
		return p.Changes[i].ID.String() < p.Changes[j].ID.String()
	})
	return p, nil
}

// Apply always builds a fresh plan; callers cannot apply stale or edited plans.
// The caller must hold the state database's process guard.
func (r *Reconciler) Apply(ctx context.Context, desired []resource.Resource, revision string) (*Plan, error) {
	p, err := r.Plan(ctx, desired, revision)
	if err != nil {
		return nil, err
	}
	for _, change := range p.Changes {
		if err = r.applyChange(ctx, change, revision); err != nil {
			return p, err
		}
	}
	if err = r.Store.SetRevision(ctx, revision); err != nil {
		return p, fault.Wrap(fault.Persistence, err)
	}
	return p, nil
}

func (r *Reconciler) applyChange(ctx context.Context, c Change, revision string) (err error) {
	operation := ""
	if c.Action != resource.Noop || c.Playbook != "" {
		action := string(c.Action)
		if c.Playbook != "" {
			action = "transition"
		}
		operation, err = r.Store.BeginOperation(ctx, c.ID.String(), action)
		if err != nil {
			return fault.Wrap(fault.Persistence, err)
		}
		defer func() {
			status, code := "succeeded", ""
			if err != nil {
				status = "failed"
				code = fault.Category(err)
			}
			endErr := r.Store.EndOperation(context.WithoutCancel(ctx), operation, status, code)
			err = errors.Join(err, endErr)
		}()
		if c.Playbook != "" {
			recorder := &recordingRunner{runner: r.Runner, store: r.Store, record: transitionRecord{Execution: c.Resume, Playbook: c.Playbook, Resource: c.ID, Desired: c.After, Prune: c.Prune}}
			_, err = recorder.run(ctx, r.Playbooks[c.Playbook], c.Resume)

		} else {
			err = c.target.Apply(ctx, c.Action)
		}
		if err != nil {
			return fault.Wrap(fault.Apply, err, "module", c.ID.Module, "resource_id", c.ID.String(), "operation_id", operation)
		}
	}
	observed, err := c.target.Observe(ctx)
	if err != nil {
		return fault.Wrap(fault.Verification, err, "resource_id", c.ID.String(), "operation_id", operation)
	}
	if observed != c.After {
		return fault.New(fault.Verification, "resource did not reach desired state", "resource_id", c.ID.String(), "operation_id", operation)
	}
	if c.Prune || !c.After.Exists {
		err = r.Store.Forget(ctx, c.ID)
	} else {
		err = r.Store.SaveManaged(ctx, state.Managed{ID: c.ID, Applied: c.After, Revision: revision})
	}
	if err != nil {
		return fault.Wrap(fault.Persistence, err, "resource_id", c.ID.String())
	}
	if c.Playbook != "" {
		key := transitionKey(c.ID, c.After)
		if err = r.Store.SetMetadata(ctx, key, ""); err != nil {
			return fault.Wrap(fault.Persistence, err)
		}
	}
	return nil
}

// A runner can implement PreparedRunner so its instance ID is checkpointed before
// the first side effect; the built-in DAG engine implements it.
type (
	PreparedRunner interface {
		Prepare(context.Context, *playbook.Playbook) (string, error)
	}
	recordingRunner struct {
		runner PlaybookRunner
		store  Repository
		record transitionRecord
	}
)

func (r *recordingRunner) run(ctx context.Context, p *playbook.Playbook, resume string) (string, error) {
	if resume == "" {
		prepared, ok := r.runner.(PreparedRunner)
		if !ok {
			return "", fault.New(fault.Validation, "transition runner must support durable preparation")
		}
		id, err := prepared.Prepare(ctx, p)
		if err != nil {
			return "", err
		}
		r.record.Execution = id
		encoded, err := json.Marshal(r.record)
		if err != nil {
			return id, fault.Wrap(fault.Persistence, err)
		}
		if err = r.store.SetMetadata(ctx, transitionKey(r.record.Resource, r.record.Desired), string(encoded)); err != nil {
			return id, err
		}
		resume = id
	}
	return r.runner.Run(ctx, p, resume)
}

// A transition intent remains pending until both its DAG and final resource
// verification have succeeded, even when the resource already looks satisfied.
type transitionRecord struct {
	Execution string            `json:"execution_id"`
	Playbook  string            `json:"playbook"`
	Resource  resource.ID       `json:"resource"`
	Desired   resource.Snapshot `json:"desired"`
	Prune     bool              `json:"prune,omitempty"`
}

func transitionKey(id resource.ID, desired resource.Snapshot) string {
	return fmt.Sprintf("transition:%s:%s", id.String(), desired.Fingerprint())
}

func (r *Reconciler) pendingTransitions(ctx context.Context, desired []resource.Resource) (map[resource.ID]transitionRecord, error) {
	values, err := r.Store.MetadataByPrefix(ctx, "transition:")
	if err != nil {
		return nil, fault.Wrap(fault.Persistence, err)
	}
	targets := map[resource.ID]resource.Snapshot{}
	for _, v := range desired {
		targets[v.ID()] = v.Desired()
	}
	pending := map[resource.ID]transitionRecord{}
	for key, value := range values {
		if value == "" {
			continue
		}
		var record transitionRecord
		if err = json.Unmarshal([]byte(value), &record); err != nil || record.Execution == "" || record.Playbook == "" || record.Resource.Validate() != nil || transitionKey(record.Resource, record.Desired) != key {
			return nil, fault.New(fault.Recovery, "invalid transition recovery metadata")
		}
		want, present := targets[record.Resource]
		if !present && record.Prune {
			handle, restoreErr := r.Registry.Restore(record.Resource)
			if restoreErr != nil {
				return nil, restoreErr
			}
			want = handle.Desired()
			present = true
		}
		if !present || want != record.Desired {
			return nil, fault.New(fault.Recovery, "resume the unfinished transition before changing or removing its desired resource", "module", record.Resource.Module, "resource_id", record.Resource.String(), "operation_id", record.Execution)
		}
		if _, exists := pending[record.Resource]; exists {
			return nil, fault.New(fault.Recovery, "multiple unfinished transitions for one resource", "resource_id", record.Resource.String())
		}
		pending[record.Resource] = record
	}
	return pending, nil
}
