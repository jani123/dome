package registry

import (
	"fmt"
	"sort"

	"github.com/jani123/dome/core/fault"
	"github.com/jani123/dome/core/resource"
)

// Registry maps explicitly supplied module names to restoration implementations.
type Registry struct{ modules map[string]resource.Module }

// New constructs an explicit registry and rejects nil or duplicate modules.
func New(modules ...resource.Module) (*Registry, error) {
	r := &Registry{modules: map[string]resource.Module{}}
	for _, m := range modules {
		if m == nil {
			return nil, fault.New(fault.Validation, "module name is required")
		}
		if err := (resource.ID{Module: m.Name(), Name: "registration"}).Validate(); err != nil {
			return nil, fault.Wrap(fault.Validation, err)
		}
		if _, ok := r.modules[m.Name()]; ok {
			return nil, fault.New(fault.Validation, "duplicate module", "module", m.Name())
		}
		r.modules[m.Name()] = m
	}
	return r, nil
}

// Names returns registered module names in sorted order.
func (r *Registry) Names() []string {
	names := []string{}
	for n := range r.modules {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// Has reports whether a module is registered.
func (r *Registry) Has(name string) bool { _, ok := r.modules[name]; return ok }

// Restore asks the owning module for a validated deletion handle.
func (r *Registry) Restore(id resource.ID) (resource.Resource, error) {
	m, ok := r.modules[id.Module]
	if !ok {
		return nil, fault.Wrap(fault.Validation, fmt.Errorf("module is not registered"), "module", id.Module)
	}
	v, err := m.Restore(id)
	if err != nil {
		return nil, fault.Wrap(fault.Validation, err, "module", id.Module, "resource_id", id.String())
	}
	if v == nil || v.ID() != id {
		return nil, fault.New(fault.Validation, "module restored the wrong resource identity", "module", id.Module, "resource_id", id.String())
	}
	if err = v.Validate(); err != nil {
		return nil, fault.Wrap(fault.Validation, err, "module", id.Module, "resource_id", id.String())
	}
	return v, nil
}
