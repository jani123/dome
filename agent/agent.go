// Package agent composes explicitly registered modules, typed desired state, and durable playbooks.
package agent

import (
	"context"
	"fmt"
	"sort"

	"github.com/jani123/dome/cli"
	"github.com/jani123/dome/config"
	"github.com/jani123/dome/core/dag"
	"github.com/jani123/dome/core/fault"
	"github.com/jani123/dome/core/playbook"
	"github.com/jani123/dome/core/reconcile"
	"github.com/jani123/dome/core/registry"
	"github.com/jani123/dome/core/resource"
	"github.com/jani123/dome/core/state"
	"github.com/spf13/cobra"
)

type (
	DesiredState func(context.Context, config.Document) ([]resource.Resource, error)
	Agent        struct {
		modules     []resource.Module
		registry    *registry.Registry
		books       map[string]*playbook.Playbook
		desired     DesiredState
		source      config.Source
		sqlite      string
		version     string
		concurrency int
	}
	Option func(*Agent) error
)

func WithModules(modules ...resource.Module) Option {
	return func(a *Agent) error { a.modules = append(a.modules, modules...); return nil }
}

func WithSQLite(path string) Option {
	return func(a *Agent) error {
		if path == "" {
			return fault.New(fault.Validation, "SQLite path cannot be empty")
		}
		a.sqlite = path
		return nil
	}
}

func WithDesiredState(fn DesiredState) Option {
	return func(a *Agent) error { a.desired = fn; return nil }
}

func WithSource(source config.Source) Option {
	return func(a *Agent) error { a.source = source; return nil }
}

func WithVersion(version string) Option {
	return func(a *Agent) error { a.version = version; return nil }
}

func WithConcurrency(n int) Option {
	return func(a *Agent) error {
		if n < 1 {
			return fault.New(fault.Validation, "concurrency must be positive")
		}
		a.concurrency = n
		return nil
	}
}

func WithPlaybooks(books ...*playbook.Playbook) Option {
	return func(a *Agent) error {
		for _, book := range books {
			if err := book.Validate(); err != nil {
				return fault.Wrap(fault.Validation, err)
			}
			if _, ok := a.books[book.Name]; ok {
				return fault.New(fault.Validation, "duplicate playbook")
			}
			a.books[book.Name] = book
		}
		return nil
	}
}

func New(options ...Option) (*Agent, error) {
	a := &Agent{books: map[string]*playbook.Playbook{}, source: config.Local{}, sqlite: ".dome/state.db", version: "dev", concurrency: 4}
	for _, option := range options {
		if option == nil {
			return nil, fault.New(fault.Validation, "nil agent option")
		}
		if err := option(a); err != nil {
			return nil, err
		}
	}
	r, err := registry.New(a.modules...)
	if err != nil {
		return nil, err
	}
	a.registry = r
	return a, nil
}
func (a *Agent) Command() *cobra.Command           { return cli.New(a) }
func (a *Agent) Execute(ctx context.Context) error { return a.Command().ExecuteContext(ctx) }
func (a *Agent) Modules() []string                 { return a.registry.Names() }
func (a *Agent) Playbooks() []string {
	names := []string{}
	for name := range a.books {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
func (a *Agent) Version() string { return a.version }
func (a *Agent) open(ctx context.Context, o cli.Options) (*state.Store, func(), error) {
	path := o.StatePath
	if path == "" {
		path = a.sqlite
	}
	s, err := state.Open(path)
	if err != nil {
		return nil, nil, fault.Wrap(fault.Persistence, err)
	}
	unlock, err := s.Guard(ctx)
	if err != nil {
		s.Close()
		return nil, nil, err
	}
	return s, func() { unlock(); _ = s.Close() }, nil
}

func (a *Agent) evaluate(ctx context.Context, o cli.Options) ([]resource.Resource, string, error) {
	if a.desired == nil {
		return nil, "", fault.New(fault.Validation, "desired-state function is required for reconciliation")
	}
	source := a.source
	if o.Git.URL != "" {
		o.Git.SecretsPath = o.SecretsPath
		source = config.Git{Options: o.Git}
	} else if o.LocalPath != "" || o.SecretsPath != "" {
		source = config.Local{Path: o.LocalPath, SecretsPath: o.SecretsPath}
	}
	if source == nil {
		return nil, "", fault.New(fault.Validation, "configuration source is nil")
	}
	doc, err := source.Load(ctx)
	if err != nil {
		return nil, "", fault.Wrap(fault.Configuration, err)
	}
	resources, err := a.desired(ctx, doc)
	if err != nil {
		return nil, "", fault.Wrap(fault.Validation, err)
	}
	return resources, doc.Revision, nil
}

func (a *Agent) reconciler(s *state.Store) *reconcile.Reconciler {
	return &reconcile.Reconciler{Store: s, Registry: a.registry, Runner: &dag.Engine{Store: s, Concurrency: a.concurrency}, Playbooks: a.books}
}

func (a *Agent) Plan(ctx context.Context, o cli.Options) (*reconcile.Plan, error) {
	if a.desired == nil {
		return nil, fault.New(fault.Validation, "desired-state function is required for reconciliation")
	}
	s, close, err := a.open(ctx, o)
	if err != nil {
		return nil, err
	}
	defer close()
	resources, revision, err := a.evaluate(ctx, o)
	if err != nil {
		return nil, err
	}
	return a.reconciler(s).Plan(ctx, resources, revision)
}

func (a *Agent) Apply(ctx context.Context, o cli.Options) (*reconcile.Plan, error) {
	if a.desired == nil {
		return nil, fault.New(fault.Validation, "desired-state function is required for reconciliation")
	}
	s, close, err := a.open(ctx, o)
	if err != nil {
		return nil, err
	}
	defer close()
	resources, revision, err := a.evaluate(ctx, o)
	if err != nil {
		return nil, err
	}
	return a.reconciler(s).Apply(ctx, resources, revision)
}

func (a *Agent) Status(ctx context.Context, o cli.Options) (any, error) {
	s, close, err := a.open(ctx, o)
	if err != nil {
		return nil, err
	}
	defer close()
	managed, err := s.Managed(ctx)
	if err != nil {
		return nil, err
	}
	revision, err := s.Revision(ctx)
	if err != nil {
		return nil, err
	}
	type item struct {
		state.Managed
		Observed resource.Snapshot `json:"observed"`
	}
	items := []item{}
	for _, m := range managed {
		handle, err := a.registry.Restore(m.ID)
		if err != nil {
			return nil, err
		}
		observed, err := handle.Observe(ctx)
		if err != nil {
			return nil, fault.Wrap(fault.Observation, err, "resource_id", m.ID.String())
		}
		items = append(items, item{m, observed})
	}
	return struct {
		Revision  string `json:"revision"`
		Resources []item `json:"resources"`
	}{revision, items}, nil
}

func (a *Agent) RunPlaybook(ctx context.Context, o cli.Options, name, resume string) (string, error) {
	book, ok := a.books[name]
	if !ok {
		return "", fault.New(fault.Validation, "playbook is not registered")
	}
	s, close, err := a.open(ctx, o)
	if err != nil {
		return "", err
	}
	defer close()
	if a.desired != nil {
		if _, _, err = a.evaluate(ctx, o); err != nil {
			return "", err
		}
	}
	engine := &dag.Engine{Store: s, Concurrency: a.concurrency}
	return engine.Run(ctx, book, resume)
}

func (a *Agent) Operations(ctx context.Context, o cli.Options) ([]state.Operation, error) {
	s, close, err := a.open(ctx, o)
	if err != nil {
		return nil, err
	}
	defer close()
	ops, err := s.Operations(ctx)
	if err != nil {
		return nil, err
	}
	executions, err := s.Executions(ctx)
	if err != nil {
		return nil, err
	}
	for _, e := range executions {
		nodes, err := s.Nodes(ctx, e.ID)
		if err != nil {
			return nil, err
		}
		ordered := []state.Node{}
		for _, n := range nodes {
			ordered = append(ordered, n)
		}
		sort.Slice(ordered, func(i, j int) bool { return ordered[i].ID < ordered[j].ID })
		attempts, err := s.Attempts(ctx, e.ID)
		if err != nil {
			return nil, err
		}
		ops = append(ops, state.Operation{ID: e.ID, Resource: e.Playbook, Action: "playbook", Status: e.Status, Nodes: ordered, Attempts: attempts})
	}
	return ops, nil
}

// ErrorText deliberately uses the public error string, never rich debug formatting.
func ErrorText(err error) string { return fmt.Sprint(err) }

var _ cli.Backend = (*Agent)(nil)
