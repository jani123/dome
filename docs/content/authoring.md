---
title: Resources and playbooks
weight: 30
---

## Resources

The starter is a clean external Go consumer. Its typed desired-state callback
returns resources such as `file.File`. Register concrete modules with
`agent.WithModules`; core packages do not import them. Module authors implement
`resource.Resource` and `resource.Module` through public packages. Never use
`init()` for global registration or load runtime plugins.

Use stable `(module, name)` identities. Validate unsafe input before changes.
Observe actual host state, decide the change, apply it, verify a fresh observation,
and only then persist applied state. Snapshots must contain non-secret data such
as hashes; they appear in plans and SQLite. `Module.Restore` must reconstruct a
validated deletion handle from identity without old secret contents.

The installed examples are executable Go documentation:

```sh
go doc github.com/jani123/dome/agent
go doc github.com/jani123/dome/core/resource
go doc github.com/jani123/dome/modules/file
```

## Include external modules

For a new project, repeat `--module` for each external root Go module:

```sh
go run github.com/jani123/dome/cmd/dome@main init my-machine \
  --project-module example.com/my-machine \
  --module github.com/acme/dome-widget@v0.1.0
```

Replace the illustrative external path with a real published module and exact
version. This initializer option is available in development builds; older tagged
initializers do not support it. Each selected root package must export `Module`.
The generated Go imports and registers `Module{}` with a unique alias; edit
`cmd/agent/main.go` to supply any required constructor, backend, or configuration
parameters before building. Initialization resolves dependencies, preserves the
requested version pins, and never compiles or applies the generated agent.

## Playbooks and transitions

Use explicit graph dependencies and named resource locks, rather than list order.
Every step requires Observe, Apply, and Verify plus `Repeatable: true` or an
explicit Recover callback. Callbacks must honor context cancellation. Increment
playbook Version when behavior changes; a changed fingerprint cannot resume an
old execution.

A resource can select a registered playbook through `resource.Transition`.
Otherwise it must explicitly declare a safe default. The same rule applies to
pruning. Keep the runtime's persisted attempts before effects and verified
resource state after effects. Do not bypass an unfinished transition by changing
or removing its target.

For a full transition demonstration and registered playbooks, inspect
`examples/basic-agent` in the matching Dome source version. That example contains
intentional failure/crash hooks; do not copy them into a production starter.
