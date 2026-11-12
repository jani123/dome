package main

import (
	"testing"

	"github.com/jani123/dome/core/resource"
	"github.com/jani123/dome/modules/file"
	"github.com/stretchr/testify/require"
)

func TestVersionTransitionAndSafePruning(t *testing.T) {
	v := &versionResource{File: file.File{Path: "/version", Content: []byte("2"), Mode: 0o600}}
	name, safe, err := v.Transition(resource.Snapshot{})
	require.NoError(t, err)
	require.Equal(t, "set-version", name)
	require.False(t, safe)
	v.Absent = true
	name, safe, err = v.Transition(resource.Snapshot{})
	require.NoError(t, err)
	require.Empty(t, name)
	require.True(t, safe)
	require.Equal(t, "demo-version", v.ID().Module)
}
