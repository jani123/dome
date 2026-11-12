package playbook_test

import (
	"context"
	"testing"

	"github.com/jani123/dome/core/playbook"
	"github.com/stretchr/testify/require"
)

func step(id string, deps ...string) *playbook.Step {
	return &playbook.Step{ID: id, Repeatable: true, Dependencies: deps, Observe: func(context.Context) (playbook.Status, error) { return playbook.NeedsApply, nil }, Apply: func(context.Context) error { return nil }, Verify: func(context.Context) error { return nil }}
}

func TestGraphValidation(t *testing.T) {
	for _, c := range []struct {
		name string
		p    *playbook.Playbook
	}{{"cycle", playbook.New("x", step("a", "b"), step("b", "a"))}, {"missing", playbook.New("x", step("a", "b"))}, {"duplicate", playbook.New("x", step("a"), step("a"))}, {"nil", playbook.New("x", nil)}, {"empty", playbook.New("", step("a"))}} {
		t.Run(c.name, func(t *testing.T) { require.Error(t, c.p.Validate()) })
	}
	require.NoError(t, playbook.New("x", step("b", "a"), step("a")).Validate())
}

func TestFingerprint(t *testing.T) {
	a := playbook.New("x", step("a"), step("b", "a"))
	b := playbook.New("x", step("b", "a"), step("a"))
	require.Equal(t, a.Fingerprint(), b.Fingerprint())
	b.Version = "2"
	require.NotEqual(t, a.Fingerprint(), b.Fingerprint())
}

func TestNonrepeatableStepsRequireRecovery(t *testing.T) {
	s := step("unsafe")
	s.Repeatable = false
	require.ErrorContains(t, playbook.New("test", s).Validate(), "must be repeatable or provide recovery")
	s.Recover = func(context.Context) error { return nil }
	require.NoError(t, playbook.New("test", s).Validate())
	branch := step("branch")
	branch.Repeatable = false
	s.OnFailure = branch
	require.Error(t, playbook.New("test", s).Validate())
}
