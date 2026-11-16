package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/jani123/dome/core/reconcile"
	"github.com/stretchr/testify/require"
)

type pollingBackend struct {
	Backend
	calls int
	err   error
}

func (b *pollingBackend) Apply(context.Context, Options) (*reconcile.Plan, error) {
	b.calls++
	return &reconcile.Plan{Revision: "polled"}, b.err
}

type pollingClock struct {
	calls      int
	durations  []time.Duration
	afterApply func(int)
}

func (c *pollingClock) Sleep(ctx context.Context, d time.Duration) error {
	c.calls++
	if c.afterApply != nil {
		c.afterApply(c.calls)
	}
	c.durations = append(c.durations, d)
	if c.calls == 3 {
		return context.Canceled
	}
	return nil
}

func TestRunPollsWithInjectedTime(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		wait time.Duration
	}{
		{"default", []string{"run", "--json"}, 30 * time.Second},
		{"override", []string{"run", "--json", "--interval", "1h"}, time.Hour},
	} {
		t.Run(tc.name, func(t *testing.T) {
			backend := &pollingBackend{}
			clock := &pollingClock{afterApply: func(n int) { require.Equal(t, n, backend.calls) }}
			cmd := newWithClock(backend, clock, WithWorkspaceStore(&memoryWorkspace{}))
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetArgs(tc.args)
			err := cmd.ExecuteContext(context.Background())
			require.ErrorIs(t, err, context.Canceled)
			require.Equal(t, 3, backend.calls)
			require.Equal(t, []time.Duration{tc.wait, tc.wait, tc.wait}, clock.durations)
			require.Contains(t, out.String(), "polled")
		})
	}
}

func TestRunRejectsNonPositiveIntervals(t *testing.T) {
	for _, value := range []string{"0", "-1s"} {
		t.Run(value, func(t *testing.T) {
			backend := &pollingBackend{}
			clock := &pollingClock{}
			cmd := newWithClock(backend, clock, WithWorkspaceStore(&memoryWorkspace{}))
			cmd.SetOut(io.Discard)
			cmd.SetArgs([]string{"run", "--interval", value})
			require.Equal(t, 2, ExitCode(cmd.Execute()))
			require.Zero(t, backend.calls)
			require.Zero(t, clock.calls)
		})
	}
}

func TestRunFailureEndsPolling(t *testing.T) {
	failure := errors.New("reconcile failed")
	backend := &pollingBackend{err: failure}
	clock := &pollingClock{}
	cmd := newWithClock(backend, clock, WithWorkspaceStore(&memoryWorkspace{}))
	cmd.SetArgs([]string{"run"})
	require.ErrorIs(t, cmd.Execute(), failure)
	require.Equal(t, 1, backend.calls)
	require.Zero(t, clock.calls)
}

func TestApplyRemainsOneShot(t *testing.T) {
	backend := &pollingBackend{}
	clock := &pollingClock{}
	cmd := newWithClock(backend, clock, WithWorkspaceStore(&memoryWorkspace{}))
	cmd.SetOut(io.Discard)
	cmd.SetArgs([]string{"apply"})
	require.NoError(t, cmd.Execute())
	require.Equal(t, 1, backend.calls)
	require.Zero(t, clock.calls)
}

func TestRunHelpDocumentsDefault(t *testing.T) {
	cmd := newWithClock(&pollingBackend{}, &pollingClock{}, WithWorkspaceStore(&memoryWorkspace{}))
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"run", "--help"})
	require.NoError(t, cmd.Execute())
	require.Contains(t, out.String(), "foreground")
	require.Contains(t, out.String(), "default 30s")
	require.Contains(t, out.String(), "must be positive")
}

type cancelPollingClock struct{}

func (*cancelPollingClock) Sleep(context.Context, time.Duration) error { return context.Canceled }
