package cli

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/jani123/dome/core/reconcile"
	"github.com/stretchr/testify/require"
)

type pollingBackend struct {
	Backend
	calls int
}

func (b *pollingBackend) Apply(context.Context, Options) (*reconcile.Plan, error) {
	b.calls++
	return &reconcile.Plan{}, nil
}

type pollingClock struct {
	calls     int
	durations []time.Duration
}

func (c *pollingClock) Sleep(ctx context.Context, d time.Duration) error {
	c.calls++
	c.durations = append(c.durations, d)
	if c.calls == 3 {
		return context.Canceled
	}
	return nil
}

func TestRunPollsWithInjectedTime(t *testing.T) {
	backend := &pollingBackend{}
	clock := &pollingClock{}
	cmd := newWithClock(backend, clock)
	cmd.SetOut(io.Discard)
	cmd.SetArgs([]string{"run", "--interval", "1h"})
	err := cmd.ExecuteContext(context.Background())
	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, 3, backend.calls)
	require.Equal(t, []time.Duration{time.Hour, time.Hour, time.Hour}, clock.durations)
}
