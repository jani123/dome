package state

import (
	"context"
	"errors"
	"testing"

	"github.com/jani123/dome/core/fault"
	"github.com/stretchr/testify/require"
)

func TestStateErrorsPreserveCausesAndCategories(t *testing.T) {
	cause := errors.New("database error")
	err := stateError(cause)
	require.ErrorIs(t, err, cause)
	require.Equal(t, string(fault.Persistence), fault.Category(err))
	conflict := fault.New(fault.Conflict, "locked")
	require.Equal(t, conflict, stateError(conflict))
	require.Nil(t, stateError(nil))
}

func TestExecutionIDs(t *testing.T) {
	a, b := NewID(), NewID()
	require.Len(t, a, 32)
	require.Len(t, b, 32)
	require.NotEqual(t, a, b)
}

func TestCancelledGuardDoesNotTouchDatabase(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	store := &Store{}
	_, err := store.Guard(ctx)
	require.ErrorIs(t, err, context.Canceled)
}
