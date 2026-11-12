// Package dag executes playbooks with persisted attempts and observation-led recovery.
package dag

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jani123/dome/core/fault"
	"github.com/jani123/dome/core/playbook"
	"github.com/jani123/dome/core/state"
)

// Repository abstracts SQLite for deterministic scheduler tests.
type Repository interface {
	NewExecution(context.Context, string, string, string) error
	Execution(context.Context, string) (state.Execution, error)
	FinishExecution(context.Context, string, string) error
	Nodes(context.Context, string) (map[string]state.Node, error)
	SaveNode(context.Context, string, state.Node) error
	Acquire(context.Context, []string, string) (bool, error)
	Release(context.Context, string) error
}
type (
	// Clock isolates retry waits and callback deadlines for deterministic tests.
	Clock interface {
		Sleep(context.Context, time.Duration) error
		WithTimeout(context.Context, time.Duration) (context.Context, context.CancelFunc)
	}
	// RealClock uses cancellable host timers.
	RealClock struct{}
)

// Sleep waits with cancellation support.
func (RealClock) Sleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// Engine schedules persisted workflows; callers must hold the store process guard.
type Engine struct {
	Store       Repository
	Concurrency int
	Clock       Clock
}

// Run creates a new instance when resume is empty. The caller must hold its
// database's process guard for the duration of this call.
func (e *Engine) Run(ctx context.Context, p *playbook.Playbook, resume string) (string, error) {
	if err := p.Validate(); err != nil {
		return "", fault.Wrap(fault.Validation, err)
	}
	id := resume
	if id == "" {
		id = state.NewID()
		if err := e.Store.NewExecution(ctx, id, p.Name, p.Fingerprint()); err != nil {
			return "", fault.Wrap(fault.Persistence, err)
		}
	} else {
		v, err := e.Store.Execution(ctx, id)
		if err != nil {
			return id, fault.Wrap(fault.Persistence, err)
		}
		if v.Playbook != p.Name || v.Fingerprint != p.Fingerprint() {
			return id, fault.New(fault.Validation, "cannot resume a changed playbook", "operation_id", id)
		}
	}
	nodes, err := e.Store.Nodes(ctx, id)
	if err != nil {
		return id, err
	}
	// Completed nodes are re-observed in dependency order, not trusted from SQLite.
	done := map[string]bool{}
	failed := map[string]bool{}
	pending := map[string]*playbook.Step{}
	for _, s := range p.Steps {
		pending[s.ID] = s
	}
	limit := e.Concurrency
	if limit <= 0 {
		limit = 4
	}
	type completion struct {
		id  string
		err error
	}
	results := make(chan completion, limit)
	active := 0
	var all []error
	for len(pending) > 0 || active > 0 {
		blocked := false
		if ctx.Err() == nil {
			for _, s := range p.Steps {
				if _, ok := pending[s.ID]; !ok {
					continue
				}
				ready := true
				bad := false
				for _, dependency := range s.Dependencies {
					if failed[dependency] {
						bad = true
					}
					if !done[dependency] {
						ready = false
					}
				}
				if bad {
					failed[s.ID] = true
					delete(pending, s.ID)
					blocked = true
					if err = e.Store.SaveNode(ctx, id, state.Node{ID: s.ID, Status: "blocked", Attempts: nodes[s.ID].Attempts}); err != nil {
						all = append(all, err)
					}
					continue
				}
				if ready && active < limit {
					delete(pending, s.ID)
					active++
					go func(s *playbook.Step) { results <- completion{s.ID, e.runStep(ctx, id, s, s.ID, nodes[s.ID])} }(s)
				}
			}
		}
		if active == 0 {
			if blocked {
				continue
			}
			break
		}
		// Drain active callbacks on cancellation; callbacks must honor their context.
		result := <-results
		active--
		if result.err != nil {
			failed[result.id] = true
			all = append(all, result.err)
		} else {
			done[result.id] = true
		}
	}

	if ctx.Err() != nil {
		all = append(all, ctx.Err())
	}
	status := "succeeded"
	if len(all) > 0 {
		status = "failed"
	}
	if ctx.Err() != nil {
		status = "interrupted"
	}
	if err = e.Store.FinishExecution(context.WithoutCancel(ctx), id, status); err != nil {
		all = append(all, err)
	}
	return id, errors.Join(all...)
}

func (e *Engine) runStep(ctx context.Context, execution string, s *playbook.Step, nodeID string, old state.Node) (err error) {
	defer func() {
		if err != nil {
			err = fault.Wrap(fault.Apply, err, "module", "dag", "operation_id", execution, "dag_node_id", nodeID)
		}
	}()
	clock := e.Clock
	if clock == nil {
		clock = RealClock{}
	}
	owner := execution + ":" + nodeID
	for {
		ok, lockErr := e.Store.Acquire(ctx, s.Locks, owner)
		if lockErr != nil {
			return lockErr
		}
		if ok {
			break
		}
		if err = clock.Sleep(ctx, 10*time.Millisecond); err != nil {
			return err
		}
	}
	defer func() { err = errors.Join(err, e.Store.Release(context.WithoutCancel(ctx), owner)) }()
	attempts := s.Retry.Attempts
	if attempts <= 0 {
		attempts = 1
	}
	node := old
	node.ID = nodeID
	for attempt := 0; attempt < attempts; attempt++ {
		if err = ctx.Err(); err != nil {
			return err
		}
		stepCtx := ctx
		cancel := func() {}
		if s.Timeout > 0 {
			stepCtx, cancel = clock.WithTimeout(ctx, s.Timeout)
		}
		err = e.attempt(stepCtx, execution, s, &node)
		cancel()
		if err == nil {
			node.Status = "succeeded"
			node.Code = ""
			return e.Store.SaveNode(ctx, execution, node)
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		node.Status = "failed"
		node.Code = fault.Category(err)
		if node.Code == "internal" {
			node.Code = string(fault.Apply)
		}
		if saveErr := e.Store.SaveNode(ctx, execution, node); saveErr != nil {
			return errors.Join(err, saveErr)
		}
		if attempt+1 < attempts {
			if sleepErr := clock.Sleep(ctx, s.Retry.Delay); sleepErr != nil {
				return errors.Join(err, sleepErr)
			}
		}
	}
	if s.OnFailure != nil {
		recoveryID := nodeID + "/recovery"
		nodes, loadErr := e.Store.Nodes(ctx, execution)
		if loadErr != nil {
			return errors.Join(err, loadErr)
		} // Release locks before entering an explicit recovery branch.
		if releaseErr := e.Store.Release(ctx, owner); releaseErr != nil {
			return errors.Join(err, releaseErr)
		}
		recoveryErr := e.runStep(ctx, execution, s.OnFailure, recoveryID, nodes[recoveryID])
		if recoveryErr != nil {
			return errors.Join(err, recoveryErr)
		}
	}
	return err
}

func (e *Engine) attempt(ctx context.Context, execution string, s *playbook.Step, node *state.Node) error {
	observed, err := s.Observe(ctx)
	if err != nil {
		return fault.Wrap(fault.Observation, err)
	}
	if observed == playbook.Satisfied {
		return verify(ctx, s)
	}
	if observed == playbook.InProgress {
		return fault.New(fault.Recovery, "operation is still in progress")
	}
	if observed != playbook.NeedsApply && observed != playbook.Unknown && observed != playbook.Failed {
		return fmt.Errorf("invalid observed step status")
	}
	interrupted := node.Attempts > 0 || node.Status == "recovering" || node.Status == "running"
	if observed == playbook.Unknown || observed == playbook.Failed || (interrupted && !s.Repeatable) {
		if s.Recover == nil {
			return fault.New(fault.Recovery, "step requires explicit recovery")
		}
		// Recovery is also a side effect, so checkpoint it first.
		node.Status = "recovering"
		if err = e.Store.SaveNode(ctx, execution, *node); err != nil {
			return err
		}
		if err = s.Recover(ctx); err != nil {
			return fault.Wrap(fault.Recovery, err)
		}
		observed, err = s.Observe(ctx)
		if err != nil {
			return fault.Wrap(fault.Observation, err)
		}
		if observed == playbook.Satisfied {
			return verify(ctx, s)
		}
		if observed != playbook.NeedsApply {
			return fault.New(fault.Recovery, "recovery did not establish a safe state")
		}
	}
	node.Attempts++
	node.Status = "running"
	node.Code = ""
	if err = e.Store.SaveNode(ctx, execution, *node); err != nil {
		return err
	}
	if err = s.Apply(ctx); err != nil {
		return err
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	return verify(ctx, s)
}

func (e *Engine) Prepare(ctx context.Context, p *playbook.Playbook) (string, error) {
	if err := p.Validate(); err != nil {
		return "", fault.Wrap(fault.Validation, err)
	}
	id := state.NewID()
	return id, e.Store.NewExecution(ctx, id, p.Name, p.Fingerprint())
}

func verify(ctx context.Context, s *playbook.Step) error {
	if err := s.Verify(ctx); err != nil {
		return fault.Wrap(fault.Verification, err)
	}
	return ctx.Err()
}

// WithTimeout derives a cancellable deadline using the host clock.
func (RealClock) WithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}
