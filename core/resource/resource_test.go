package resource_test

import (
	"testing"

	"github.com/jani123/dome/core/resource"
	"github.com/stretchr/testify/assert"
)

func TestDifference(t *testing.T) {
	a := resource.Snapshot{Exists: true, Value: "a"}
	b := resource.Snapshot{Exists: true, Value: "b"}
	for _, c := range []struct {
		actual, desired resource.Snapshot
		want            resource.Action
	}{{resource.Snapshot{}, a, resource.Create}, {a, b, resource.Update}, {a, resource.Snapshot{}, resource.Delete}, {a, a, resource.Noop}, {resource.Snapshot{}, resource.Snapshot{}, resource.Noop}} {
		assert.Equal(t, c.want, resource.Difference(c.actual, c.desired))
	}
	assert.Error(t, (resource.ID{}).Validate())
	assert.NoError(t, (resource.ID{Module: "file", Name: "/x"}).Validate())
}

func TestSnapshotFingerprint(t *testing.T) {
	a := resource.Snapshot{Exists: true, Value: "secret"}
	b := resource.Snapshot{Exists: true, Value: "other"}
	assert.NotContains(t, a.Fingerprint(), "secret")
	assert.Equal(t, a.Fingerprint(), a.Fingerprint())
	assert.NotEqual(t, a.Fingerprint(), b.Fingerprint())
	assert.NotEqual(t, a.Fingerprint(), (resource.Snapshot{Value: "secret"}).Fingerprint())
}
