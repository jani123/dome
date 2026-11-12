//go:build integration

package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMarkerLifecycle(t *testing.T) {
	t.Setenv("DOME_ROOT", t.TempDir())
	s := markerStep("test", "marker", 0)
	observed, err := s.Observe(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, "needs-apply", observed)
	require.NoError(t, s.Apply(context.Background()))
	require.NoError(t, s.Verify(context.Background()))
	observed, err = s.Observe(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, "satisfied", observed)
}
