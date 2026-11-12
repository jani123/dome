package dag_test

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jani123/dome/core/dag"
	"github.com/jani123/dome/core/fault"
	"github.com/jani123/dome/core/playbook"
	"github.com/jani123/dome/core/state"
	"github.com/jani123/dome/internal/testutil"
	"github.com/stretchr/testify/require"
)

func step(id string, done *atomic.Bool, calls *atomic.Int32) *playbook.Step {
	return &playbook.Step{ID: id, Repeatable: true, Observe: func(context.Context) (playbook.Status, error) {
		if done.Load() {
			return playbook.Satisfied, nil
		}
		return playbook.NeedsApply, nil
	}, Apply: func(context.Context) error { calls.Add(1); done.Store(true); return nil }, Verify: func(context.Context) error {
		if !done.Load() {
			return errors.New("not done")
		}
		return nil
	}}
}

func TestParallelDependenciesAndReobservation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	store := testutil.NewMemory()
	engine := &dag.Engine{Store: store}
	started := make(chan string, 2)
	release := make(chan struct{})
	var a, b, joined atomic.Bool
	var calls atomic.Int32
	left := step("left", &a, &calls)
	right := step("right", &b, &calls)
	for _, s := range []*playbook.Step{left, right} {
		original := s.Apply
		name := s.ID
		s.Apply = func(ctx context.Context) error {
			started <- name
			select {
			case <-release:
			case <-ctx.Done():
				return ctx.Err()
			}
			return original(ctx)
		}
	}
	join := step("join", &joined, &calls)
	join.Dependencies = []string{"left", "right"}
	join.Apply = func(context.Context) error {
		if !a.Load() || !b.Load() {
			return errors.New("dependency violated")
		}
		joined.Store(true)
		return nil
	}
	book := playbook.New("parallel", join, left, right)
	result := make(chan error, 1)
	var id string
	go func() { var err error; id, err = engine.Run(ctx, book, ""); result <- err }()
	for range 2 {
		select {
		case <-started:
		case <-ctx.Done():
			t.Fatal("independent nodes did not start concurrently")
		}
	}
	close(release)
	require.NoError(t, <-result)
	require.EqualValues(t, 2, calls.Load())
	_, err := engine.Run(ctx, book, id)
	require.NoError(t, err)
	require.EqualValues(t, 2, calls.Load())
	a.Store(false)
	_, err = engine.Run(ctx, book, id)
	require.NoError(t, err)
	require.EqualValues(t, 3, calls.Load())
}

func TestFailureBlocksOnlyDependents(t *testing.T) {
	store := testutil.NewMemory()
	engine := &dag.Engine{Store: store}
	var a, b, c atomic.Bool
	var calls atomic.Int32
	bad := step("bad", &a, &calls)
	bad.Apply = func(context.Context) error { return errors.New("failed") }
	dependent := step("dependent", &b, &calls)
	dependent.Dependencies = []string{"bad"}
	independent := step("independent", &c, &calls)
	id, err := engine.Run(context.Background(), playbook.New("test", bad, dependent, independent), "")
	require.Error(t, err)
	require.False(t, b.Load())
	require.True(t, c.Load())
	nodes, err := store.Nodes(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, "blocked", nodes["dependent"].Status)
}

func TestCrashAtExternalSuccessBeforeCheckpoint(t *testing.T) {
	ctx := context.Background()
	store := testutil.NewMemory()
	var done atomic.Bool
	var calls atomic.Int32
	book := playbook.New("crash", step("external", &done, &calls))
	store.BeforeSaveNode = func(n state.Node) error {
		if n.Status == "succeeded" {
			return errors.New("crash before completion checkpoint")
		}
		return nil
	}
	engine := &dag.Engine{Store: store}
	id, err := engine.Run(ctx, book, "")
	require.Error(t, err)
	require.True(t, done.Load())
	nodes, _ := store.Nodes(ctx, id)
	require.Equal(t, "running", nodes["external"].Status)
	store.BeforeSaveNode = nil
	_, err = engine.Run(ctx, book, id)
	require.NoError(t, err)
	require.EqualValues(t, 1, calls.Load())
	nodes, _ = store.Nodes(ctx, id)
	require.Equal(t, "succeeded", nodes["external"].Status)
}

func TestNonrepeatableRecoveryAndUnknownStates(t *testing.T) {
	for _, recoverable := range []bool{false, true} {
		t.Run(map[bool]string{false: "reject", true: "recover"}[recoverable], func(t *testing.T) {
			ctx := context.Background()
			store := testutil.NewMemory()
			var done atomic.Bool
			var calls atomic.Int32
			s := step("unsafe", &done, &calls)
			s.Repeatable = false
			recoveries := 0
			if recoverable {
				s.Recover = func(context.Context) error { recoveries++; return nil }
			}
			book := playbook.New("test", s)
			engine := &dag.Engine{Store: store}
			id, err := engine.Prepare(ctx, book)
			if !recoverable {
				require.Error(t, err)
				require.Zero(t, calls.Load())
				return
			}
			require.NoError(t, err)
			require.NoError(t, store.SaveNode(ctx, id, state.Node{ID: "unsafe", Status: "running", Attempts: 1}))
			_, err = engine.Run(ctx, book, id)
			if recoverable {
				require.NoError(t, err)
				require.Equal(t, 1, recoveries)
				require.EqualValues(t, 1, calls.Load())
			} else {
				require.Error(t, err)
				require.Zero(t, calls.Load())
			}
		})
	}
}

type fakeClock struct {
	calls  atomic.Int32
	expire bool
}

func (c *fakeClock) Sleep(ctx context.Context, d time.Duration) error {
	c.calls.Add(1)
	runtime.Gosched()
	return ctx.Err()
}

func (c *fakeClock) WithTimeout(ctx context.Context, _ time.Duration) (context.Context, context.CancelFunc) {
	child, cancel := context.WithCancel(ctx)
	if c.expire {
		cancel()
	}
	return child, cancel
}

func TestRetryAndRecoveryBranch(t *testing.T) {
	ctx := context.Background()
	store := testutil.NewMemory()
	clock := &fakeClock{}
	engine := &dag.Engine{Store: store, Clock: clock}
	var done, recovered atomic.Bool
	var calls, rcalls atomic.Int32
	s := step("retry", &done, &calls)
	s.Retry = playbook.RetryPolicy{Attempts: 2, Delay: time.Hour}
	s.Apply = func(context.Context) error { calls.Add(1); return errors.New("failed") }
	s.OnFailure = step("repair", &recovered, &rcalls)
	id, err := engine.Run(ctx, playbook.New("test", s), "")
	require.Error(t, err)
	require.EqualValues(t, 2, calls.Load())
	require.EqualValues(t, 1, clock.calls.Load())
	require.True(t, recovered.Load())
	nodes, _ := store.Nodes(ctx, id)
	require.Equal(t, "succeeded", nodes["retry/recovery"].Status)
}

type lockProbe struct {
	*testutil.Memory
	contended chan struct{}
	once      sync.Once
}

func (s *lockProbe) Acquire(ctx context.Context, names []string, owner string) (bool, error) {
	ok, err := s.Memory.Acquire(ctx, names, owner)
	if err == nil && !ok {
		s.once.Do(func() { close(s.contended) })
	}
	return ok, err
}

func TestLocksPreventOverlap(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	store := &lockProbe{Memory: testutil.NewMemory(), contended: make(chan struct{})}
	engine := &dag.Engine{Store: store, Clock: &fakeClock{}}
	var active, max atomic.Int32
	makeStep := func(id string) *playbook.Step {
		var done atomic.Bool
		var calls atomic.Int32
		s := step(id, &done, &calls)
		s.Locks = []string{"shared-resource"}
		s.Apply = func(ctx context.Context) error {
			v := active.Add(1)
			for {
				old := max.Load()
				if v <= old || max.CompareAndSwap(old, v) {
					break
				}
			}
			select {
			case <-store.contended:
			case <-ctx.Done():
				return ctx.Err()
			}
			active.Add(-1)
			done.Store(true)
			return nil
		}
		return s
	}
	_, err := engine.Run(ctx, playbook.New("locks", makeStep("a"), makeStep("b")), "")
	require.NoError(t, err)
	require.EqualValues(t, 1, max.Load())
}

func TestTimeoutAndChangedGraph(t *testing.T) {
	store := testutil.NewMemory()
	engine := &dag.Engine{Store: store, Clock: &fakeClock{expire: true}}
	var done atomic.Bool
	var calls atomic.Int32
	s := step("slow", &done, &calls)
	s.Timeout = time.Millisecond
	s.Apply = func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() }
	book := playbook.New("timeout", s)
	id, err := engine.Run(context.Background(), book, "")
	require.ErrorIs(t, err, context.Canceled)
	book.Version = "2"
	_, err = engine.Run(context.Background(), book, id)
	require.ErrorContains(t, err, "changed playbook")
}

func TestReadyDependentRunsWhileUnrelatedRootIsStillRunning(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	store := testutil.NewMemory()
	engine := &dag.Engine{Store: store, Concurrency: 2}
	var leftDone, rightDone, leafDone atomic.Bool
	var calls atomic.Int32
	left := step("left", &leftDone, &calls)
	right := step("right", &rightDone, &calls)
	leaf := step("leaf", &leafDone, &calls)
	leaf.Dependencies = []string{"left"}
	leafStarted := make(chan struct{})
	right.Apply = func(ctx context.Context) error {
		select {
		case <-leafStarted:
			rightDone.Store(true)
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	leaf.Apply = func(context.Context) error { leafDone.Store(true); close(leafStarted); return nil }
	_, err := engine.Run(ctx, playbook.New("rolling", left, right, leaf), "")
	require.NoError(t, err, "an unrelated active node must not block a ready dependent")
	require.True(t, rightDone.Load())
	require.True(t, leafDone.Load())
}

func TestUnknownAndInProgressNeverBlindlyApply(t *testing.T) {
	for _, status := range []playbook.Status{playbook.Unknown, playbook.InProgress, playbook.Failed} {
		t.Run(string(status), func(t *testing.T) {
			store := testutil.NewMemory()
			var applied atomic.Int32
			step := &playbook.Step{ID: "external", Repeatable: true, Observe: func(context.Context) (playbook.Status, error) { return status, nil }, Apply: func(context.Context) error { applied.Add(1); return nil }, Verify: func(context.Context) error { return nil }}
			_, err := (&dag.Engine{Store: store}).Run(context.Background(), playbook.New("safe", step), "")
			require.Error(t, err)
			require.Equal(t, "recovery", fault.Category(err))
			require.Zero(t, applied.Load())
		})
	}
}

func TestRecoveryCheckpointPrecedesRecoverySideEffect(t *testing.T) {
	sentinel := errors.New("cannot persist recovery")
	store := testutil.NewMemory()
	store.BeforeSaveNode = func(node state.Node) error {
		if node.Status == "recovering" {
			return sentinel
		}
		return nil
	}
	var recovered, applied atomic.Int32
	step := &playbook.Step{ID: "external", Repeatable: false, Observe: func(context.Context) (playbook.Status, error) { return playbook.Unknown, nil }, Recover: func(context.Context) error { recovered.Add(1); return nil }, Apply: func(context.Context) error { applied.Add(1); return nil }, Verify: func(context.Context) error { return nil }}
	_, err := (&dag.Engine{Store: store}).Run(context.Background(), playbook.New("recover", step), "")
	require.ErrorIs(t, err, sentinel)
	require.Zero(t, recovered.Load())
	require.Zero(t, applied.Load())
}

func TestSatisfiedObservationStillRequiresVerification(t *testing.T) {
	store := testutil.NewMemory()
	var applied atomic.Int32
	sentinel := errors.New("verification failed")
	step := &playbook.Step{ID: "external", Repeatable: true, Observe: func(context.Context) (playbook.Status, error) { return playbook.Satisfied, nil }, Apply: func(context.Context) error { applied.Add(1); return nil }, Verify: func(context.Context) error { return sentinel }}
	id, err := (&dag.Engine{Store: store}).Run(context.Background(), playbook.New("verify", step), "")
	require.ErrorIs(t, err, sentinel)
	require.Equal(t, "verification", fault.Category(err))
	require.Zero(t, applied.Load())
	nodes, _ := store.Nodes(context.Background(), id)
	require.Equal(t, "failed", nodes["external"].Status)
}
