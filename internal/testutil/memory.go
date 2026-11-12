// Package testutil supplies concurrency-safe repository fakes for unit tests.
package testutil

import (
	"context"
	"database/sql"
	"strings"
	"sync"

	"github.com/jani123/dome/core/resource"
	"github.com/jani123/dome/core/state"
)

type Memory struct {
	mu                sync.Mutex
	managed           map[resource.ID]state.Managed
	metadata          map[string]string
	operations        map[string]state.Operation
	executions        map[string]state.Execution
	nodes             map[string]map[string]state.Node
	locks             map[string]string
	BeforeSaveNode    func(state.Node) error
	BeforeSaveManaged func(state.Managed) error
}

func NewMemory() *Memory {
	return &Memory{managed: map[resource.ID]state.Managed{}, metadata: map[string]string{}, operations: map[string]state.Operation{}, executions: map[string]state.Execution{}, nodes: map[string]map[string]state.Node{}, locks: map[string]string{}}
}

func (m *Memory) Managed(context.Context) ([]state.Managed, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v := []state.Managed{}
	for _, x := range m.managed {
		v = append(v, x)
	}
	return v, nil
}

func (m *Memory) SaveManaged(_ context.Context, v state.Managed) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.BeforeSaveManaged != nil {
		if err := m.BeforeSaveManaged(v); err != nil {
			return err
		}
	}
	m.managed[v.ID] = v
	return nil
}

func (m *Memory) Forget(_ context.Context, id resource.ID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.managed, id)
	return nil
}

func (m *Memory) Metadata(_ context.Context, key string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.metadata[key], nil
}

func (m *Memory) SetMetadata(_ context.Context, key, v string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metadata[key] = v
	return nil
}

func (m *Memory) Revision(ctx context.Context) (string, error) { return m.Metadata(ctx, "revision") }

func (m *Memory) SetRevision(ctx context.Context, v string) error {
	return m.SetMetadata(ctx, "revision", v)
}

func (m *Memory) BeginOperation(_ context.Context, r, a string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id := state.NewID()
	m.operations[id] = state.Operation{ID: id, Resource: r, Action: a, Status: "running"}
	return id, nil
}

func (m *Memory) EndOperation(_ context.Context, id, status, code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	v := m.operations[id]
	v.Status = status
	v.Code = code
	m.operations[id] = v
	return nil
}

func (m *Memory) Operations(context.Context) ([]state.Operation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v := []state.Operation{}
	for _, x := range m.operations {
		v = append(v, x)
	}
	return v, nil
}

func (m *Memory) NewExecution(_ context.Context, id, book, fp string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.executions[id] = state.Execution{ID: id, Playbook: book, Fingerprint: fp, Status: "running"}
	return nil
}

func (m *Memory) Execution(_ context.Context, id string) (state.Execution, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.executions[id]
	if !ok {
		return v, sql.ErrNoRows
	}
	return v, nil
}

func (m *Memory) FinishExecution(_ context.Context, id, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	v := m.executions[id]
	v.Status = status
	m.executions[id] = v
	return nil
}

func (m *Memory) Nodes(_ context.Context, id string) (map[string]state.Node, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v := map[string]state.Node{}
	for k, x := range m.nodes[id] {
		v[k] = x
	}
	return v, nil
}

func (m *Memory) SaveNode(_ context.Context, id string, n state.Node) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.BeforeSaveNode != nil {
		if err := m.BeforeSaveNode(n); err != nil {
			return err
		}
	}
	if m.nodes[id] == nil {
		m.nodes[id] = map[string]state.Node{}
	}
	m.nodes[id][n.ID] = n
	return nil
}

func (m *Memory) Acquire(_ context.Context, names []string, owner string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, name := range names {
		if existing := m.locks[name]; existing != "" && existing != owner {
			return false, nil
		}
	}
	for _, name := range names {
		m.locks[name] = owner
	}
	return true, nil
}

func (m *Memory) Release(_ context.Context, owner string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for name, x := range m.locks {
		if x == owner {
			delete(m.locks, name)
		}
	}
	return nil
}

func (m *Memory) MetadataByPrefix(_ context.Context, prefix string) (map[string]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := map[string]string{}
	for key, value := range m.metadata {
		if strings.HasPrefix(key, prefix) {
			out[key] = value
		}
	}
	return out, nil
}
