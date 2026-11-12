package fault_test

import (
	"errors"
	"testing"

	"github.com/jani123/dome/core/fault"
	"github.com/samber/oops"
	"github.com/stretchr/testify/require"
)

func TestCausesAndContext(t *testing.T) {
	cause := errors.New("sentinel")
	err := fault.Wrap(fault.Apply, cause, "resource_id", "file:/a")
	require.ErrorIs(t, err, cause)
	require.Equal(t, "apply", fault.Category(err))
	var rich oops.OopsError
	require.ErrorAs(t, err, &rich)
	require.Equal(t, "file:/a", rich.Context()["resource_id"])
	require.Nil(t, fault.Wrap(fault.Apply, nil))
	require.Equal(t, "internal", fault.Category(cause))
}

func TestBoundaryContextPreservesSpecificErrorCategory(t *testing.T) {
	cause := errors.New("sentinel")
	err := fault.Wrap(fault.Verification, cause, "resource_id", "file:/example")
	err = fault.Wrap(fault.Apply, err, "operation_id", "execution", "dag_node_id", "step")
	require.ErrorIs(t, err, cause)
	require.Equal(t, "verification", fault.Category(err))
	var rich oops.OopsError
	require.ErrorAs(t, err, &rich)
	require.Equal(t, "execution", rich.Context()["operation_id"])
	require.Equal(t, "file:/example", rich.Context()["resource_id"])
}
