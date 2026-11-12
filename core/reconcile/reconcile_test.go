package reconcile_test

import (
	"context"
	"errors"
	"io/fs"
	"sync"
	"testing"

	"github.com/jani123/dome/core/dag"
	"github.com/jani123/dome/core/fault"
	"github.com/jani123/dome/core/playbook"
	"github.com/jani123/dome/core/reconcile"
	"github.com/jani123/dome/core/registry"
	"github.com/jani123/dome/core/resource"
	"github.com/jani123/dome/core/state"
	"github.com/jani123/dome/internal/testutil"
	"github.com/jani123/dome/modules/file"
	"github.com/stretchr/testify/require"
)

type disk struct {
	mu     sync.Mutex
	files  map[string][]byte
	writes int
	fail   bool
}

func (d *disk) Read(p string) ([]byte, fs.FileMode, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	b, ok := d.files[p]
	if !ok {
		return nil, 0, fs.ErrNotExist
	}
	return b, 0o600, nil
}

func (d *disk) WriteAtomic(p string, b []byte, m fs.FileMode) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.fail {
		return errors.New("write failed")
	}
	d.files[p] = append([]byte(nil), b...)
	d.writes++
	return nil
}

func (d *disk) Remove(p string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.files, p)
	return nil
}

func setup(t *testing.T) (*reconcile.Reconciler, *testutil.Memory, *disk) {
	t.Helper()
	d := &disk{files: map[string][]byte{}}
	store := testutil.NewMemory()
	reg, err := registry.New(file.Module{FS: d})
	require.NoError(t, err)
	return &reconcile.Reconciler{Store: store, Registry: reg}, store, d
}

func f(d *disk, p, content string) *file.File {
	return &file.File{Path: p, Content: []byte(content), Mode: 0o600, FS: d}
}

func TestCreateUpdateIdempotencePruneAndOwnership(t *testing.T) {
	ctx := context.Background()
	r, s, d := setup(t)
	d.files["/unrelated"] = []byte("keep")
	want := []resource.Resource{f(d, "/managed", "one")}
	p, err := r.Plan(ctx, want, "a")
	require.NoError(t, err)
	require.Equal(t, resource.Create, p.Changes[0].Action)
	require.Zero(t, d.writes)
	_, err = r.Apply(ctx, want, "a")
	require.NoError(t, err)
	p, err = r.Apply(ctx, want, "a")
	require.NoError(t, err)
	require.Equal(t, resource.Noop, p.Changes[0].Action)
	require.Equal(t, 1, d.writes)
	want[0] = f(d, "/managed", "two")
	p, err = r.Apply(ctx, want, "b")
	require.NoError(t, err)
	require.Equal(t, resource.Update, p.Changes[0].Action)
	p, err = r.Apply(ctx, nil, "c")
	require.NoError(t, err)
	require.True(t, p.Changes[0].Prune)
	require.NotContains(t, d.files, "/managed")
	require.Equal(t, []byte("keep"), d.files["/unrelated"])
	m, err := s.Managed(ctx)
	require.NoError(t, err)
	require.Empty(t, m)
	revision, err := s.Revision(ctx)
	require.NoError(t, err)
	require.Equal(t, "c", revision)
}

func TestInvalidDesiredStateNeverPrunes(t *testing.T) {
	ctx := context.Background()
	r, _, d := setup(t)
	_, err := r.Apply(ctx, []resource.Resource{f(d, "/owned", "keep")}, "a")
	require.NoError(t, err)
	_, err = r.Apply(ctx, []resource.Resource{f(d, "relative", "bad")}, "b")
	require.Error(t, err)
	require.Contains(t, d.files, "/owned")
	_, err = r.Apply(ctx, []resource.Resource{f(d, "/x", "a"), f(d, "/x", "b")}, "b")
	require.Error(t, err)
	require.Contains(t, d.files, "/owned")
}

func TestUnownedDeletionIsRejected(t *testing.T) {
	r, _, d := setup(t)
	d.files["/unrelated"] = []byte("keep")
	v := f(d, "/unrelated", "")
	v.Absent = true
	_, err := r.Apply(context.Background(), []resource.Resource{v}, "a")
	require.Error(t, err)
	require.Contains(t, d.files, "/unrelated")
}

func TestFailureDoesNotAdvanceAppliedStateOrRevision(t *testing.T) {
	ctx := context.Background()
	r, s, d := setup(t)
	_, err := r.Apply(ctx, []resource.Resource{f(d, "/x", "old")}, "old")
	require.NoError(t, err)
	d.fail = true
	_, err = r.Apply(ctx, []resource.Resource{f(d, "/x", "new")}, "new")
	require.Error(t, err)
	m, err := s.Managed(ctx)
	require.NoError(t, err)
	require.Equal(t, f(d, "/x", "old").Desired(), m[0].Applied)
	revision, err := s.Revision(ctx)
	require.NoError(t, err)
	require.Equal(t, "old", revision)
	ops, err := s.Operations(ctx)
	require.NoError(t, err)
	statuses := []string{}
	for _, op := range ops {
		statuses = append(statuses, op.Status)
	}
	require.Contains(t, statuses, "failed")
}

func TestCrashBetweenEffectAndPersistenceReobserves(t *testing.T) {
	ctx := context.Background()
	r, s, d := setup(t)
	sentinel := errors.New("checkpoint failed")
	s.BeforeSaveManaged = func(state.Managed) error { return sentinel }
	_, err := r.Apply(ctx, []resource.Resource{f(d, "/x", "value")}, "a")
	require.ErrorIs(t, err, sentinel)
	require.Equal(t, 1, d.writes)
	s.BeforeSaveManaged = nil
	p, err := r.Apply(ctx, []resource.Resource{f(d, "/x", "value")}, "a")
	require.NoError(t, err)
	require.Equal(t, resource.Noop, p.Changes[0].Action)
	require.Equal(t, 1, d.writes)
}

type transitionFile struct {
	*file.File
	unsafe bool
}

func (f *transitionFile) Transition(resource.Snapshot) (string, bool, error) {
	if f.unsafe {
		return "", false, nil
	}
	return "upgrade", false, nil
}

func TestTransitionVerificationAndFailure(t *testing.T) {
	ctx := context.Background()
	r, s, d := setup(t)
	target := &transitionFile{File: f(d, "/version", "2")}
	calls := 0
	book := playbook.New("upgrade", &playbook.Step{ID: "upgrade", Repeatable: true, Observe: func(context.Context) (playbook.Status, error) {
		if string(d.files["/version"]) == "2" {
			return playbook.Satisfied, nil
		}
		return playbook.NeedsApply, nil
	}, Apply: func(ctx context.Context) error { calls++; return target.Apply(ctx, resource.Update) }, Verify: func(context.Context) error { return nil }})
	r.Playbooks = map[string]*playbook.Playbook{"upgrade": book}
	r.Runner = &dag.Engine{Store: s}
	_, err := r.Apply(ctx, []resource.Resource{target}, "v2")
	require.NoError(t, err)
	require.Equal(t, 1, calls)
	target.Content = []byte("3")
	book.Steps[0].Observe = func(context.Context) (playbook.Status, error) { return playbook.NeedsApply, nil }
	book.Steps[0].Apply = func(context.Context) error { return errors.New("transition failed") }
	_, err = r.Apply(ctx, []resource.Resource{target}, "v3")
	require.Error(t, err)
	m, _ := s.Managed(ctx)
	require.Equal(t, "v2", m[0].Revision)
	target.unsafe = true
	_, err = r.Plan(ctx, []resource.Resource{target}, "v3")
	require.Error(t, err)
}

func TestSatisfiedResourceStillResumesUnfinishedTransition(t *testing.T) {
	ctx := context.Background()
	r, s, d := setup(t)
	target := &transitionFile{File: f(d, "/version", "2")}
	finalized := false
	fail := true
	write := &playbook.Step{ID: "write", Repeatable: true, Observe: func(ctx context.Context) (playbook.Status, error) {
		actual, err := target.Observe(ctx)
		if err != nil {
			return playbook.Unknown, err
		}
		if actual == target.Desired() {
			return playbook.Satisfied, nil
		}
		return playbook.NeedsApply, nil
	}, Apply: func(ctx context.Context) error { return target.Apply(ctx, resource.Update) }, Verify: func(ctx context.Context) error {
		actual, err := target.Observe(ctx)
		if err != nil {
			return err
		}
		if actual != target.Desired() {
			return errors.New("wrong version")
		}
		return nil
	}}
	finalize := &playbook.Step{ID: "finalize", Dependencies: []string{"write"}, Repeatable: true, Observe: func(context.Context) (playbook.Status, error) {
		if finalized {
			return playbook.Satisfied, nil
		}
		return playbook.NeedsApply, nil
	}, Apply: func(context.Context) error {
		if fail {
			return errors.New("interrupted after resource changed")
		}
		finalized = true
		return nil
	}, Verify: func(context.Context) error {
		if !finalized {
			return errors.New("not finalized")
		}
		return nil
	}}
	r.Playbooks = map[string]*playbook.Playbook{"upgrade": playbook.New("upgrade", write, finalize)}
	r.Runner = &dag.Engine{Store: s}
	_, err := r.Apply(ctx, []resource.Resource{target}, "revision-2")
	require.Error(t, err)
	require.Equal(t, []byte("2"), d.files["/version"])
	managed, _ := s.Managed(ctx)
	require.Empty(t, managed)
	fail = false
	_, err = r.Apply(ctx, []resource.Resource{target}, "revision-2")
	require.NoError(t, err)
	require.True(t, finalized, "a matching resource must not bypass the unfinished transition DAG")
	require.Equal(t, 1, d.writes, "the completed write must be observed rather than repeated")
	managed, err = s.Managed(ctx)
	require.NoError(t, err)
	require.Len(t, managed, 1)
}

func TestUnfinishedTransitionBlocksChangingOrRemovingTarget(t *testing.T) {
	ctx := context.Background()
	r, s, d := setup(t)
	target := &transitionFile{File: f(d, "/version", "2")}
	book := playbook.New("upgrade", &playbook.Step{ID: "upgrade", Repeatable: true, Observe: func(context.Context) (playbook.Status, error) { return playbook.NeedsApply, nil }, Apply: func(context.Context) error { return errors.New("interrupted") }, Verify: func(context.Context) error { return nil }})
	r.Playbooks = map[string]*playbook.Playbook{"upgrade": book}
	r.Runner = &dag.Engine{Store: s}
	_, err := r.Apply(ctx, []resource.Resource{target}, "a")
	require.Error(t, err)
	target.Content = []byte("3")
	_, err = r.Apply(ctx, []resource.Resource{target}, "b")
	require.ErrorContains(t, err, "unfinished transition")
	_, err = r.Apply(ctx, nil, "c")
	require.ErrorContains(t, err, "unfinished transition")
	require.Zero(t, d.writes)
	revision, _ := s.Revision(ctx)
	require.Empty(t, revision)
}

type pruneModule struct{ disk *disk }

func (pruneModule) Name() string { return "file" }
func (m pruneModule) Restore(id resource.ID) (resource.Resource, error) {
	v := &transitionFile{File: &file.File{Path: id.Name, Absent: true, FS: m.disk}}
	return v, v.Validate()
}

func TestPruningResumesItsWorkflowAfterTheFileWasDeleted(t *testing.T) {
	ctx := context.Background()
	r, s, d := setup(t)
	_, err := r.Apply(ctx, []resource.Resource{f(d, "/version", "1")}, "a")
	require.NoError(t, err)
	r.Registry, err = registry.New(pruneModule{disk: d})
	require.NoError(t, err)
	finalized := false
	fail := true
	deletion := &playbook.Step{ID: "delete", Repeatable: true, Observe: func(context.Context) (playbook.Status, error) {
		if _, ok := d.files["/version"]; !ok {
			return playbook.Satisfied, nil
		}
		return playbook.NeedsApply, nil
	}, Apply: func(context.Context) error { return d.Remove("/version") }, Verify: func(context.Context) error {
		if _, ok := d.files["/version"]; ok {
			return errors.New("file still exists")
		}
		return nil
	}}
	finalize := &playbook.Step{ID: "finalize", Dependencies: []string{"delete"}, Repeatable: true, Observe: func(context.Context) (playbook.Status, error) {
		if finalized {
			return playbook.Satisfied, nil
		}
		return playbook.NeedsApply, nil
	}, Apply: func(context.Context) error {
		if fail {
			return errors.New("interrupted")
		}
		finalized = true
		return nil
	}, Verify: func(context.Context) error { return nil }}
	r.Playbooks = map[string]*playbook.Playbook{"upgrade": playbook.New("upgrade", deletion, finalize)}
	r.Runner = &dag.Engine{Store: s}
	_, err = r.Apply(ctx, nil, "b")
	require.Error(t, err)
	require.NotContains(t, d.files, "/version")
	owned, _ := s.Managed(ctx)
	require.Len(t, owned, 1)
	fail = false
	_, err = r.Apply(ctx, nil, "b")
	require.NoError(t, err)
	require.True(t, finalized)
	owned, _ = s.Managed(ctx)
	require.Empty(t, owned)
}

func TestCorruptRecoveryMetadataPreventsAllResourceChanges(t *testing.T) {
	ctx := context.Background()
	r, s, d := setup(t)
	require.NoError(t, s.SetMetadata(ctx, "transition:broken", "not valid recovery metadata"))
	_, err := r.Apply(ctx, []resource.Resource{f(d, "/new", "value")}, "a")
	require.Error(t, err)
	require.Zero(t, d.writes)
}

func TestSuccessfulDAGCannotAdvanceUnverifiedResourceState(t *testing.T) {
	ctx := context.Background()
	r, s, d := setup(t)
	_, err := r.Apply(ctx, []resource.Resource{f(d, "/version", "1")}, "old")
	require.NoError(t, err)
	target := &transitionFile{File: f(d, "/version", "2")}
	book := playbook.New("upgrade", &playbook.Step{ID: "incorrect-success", Repeatable: true, Observe: func(context.Context) (playbook.Status, error) { return playbook.NeedsApply, nil }, Apply: func(context.Context) error { return nil }, Verify: func(context.Context) error { return nil }})
	r.Playbooks = map[string]*playbook.Playbook{"upgrade": book}
	r.Runner = &dag.Engine{Store: s}
	_, err = r.Apply(ctx, []resource.Resource{target}, "new")
	require.Error(t, err)
	require.Equal(t, "verification", fault.Category(err))
	managed, _ := s.Managed(ctx)
	require.Len(t, managed, 1)
	require.Equal(t, f(d, "/version", "1").Desired(), managed[0].Applied)
	revision, _ := s.Revision(ctx)
	require.Equal(t, "old", revision)
}
