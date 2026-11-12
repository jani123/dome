package registry_test

import (
	"testing"

	"github.com/jani123/dome/core/registry"
	"github.com/jani123/dome/core/resource"
	"github.com/jani123/dome/modules/file"
	"github.com/stretchr/testify/require"
)

func TestRegistration(t *testing.T) {
	r, err := registry.New(file.Module{})
	require.NoError(t, err)
	require.Equal(t, []string{"file"}, r.Names())
	v, err := r.Restore(resource.ID{Module: "file", Name: "/tmp/a"})
	require.NoError(t, err)
	require.False(t, v.Desired().Exists)
	_, err = r.Restore(resource.ID{Module: "missing", Name: "x"})
	require.Error(t, err)
	_, err = r.Restore(resource.ID{Module: "file", Name: "../x"})
	require.Error(t, err)
	_, err = registry.New(file.Module{}, file.Module{})
	require.Error(t, err)
	_, err = registry.New(nil)
	require.Error(t, err)
}
