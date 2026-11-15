---
type: Decision
title: "Project initialization"
description: "One immutable version binds generated source, dependency, skill, and template metadata."
generated: {by: codex/2026-11-13, at: "2026-11-13T07:03:33Z"}
sources:
  - id: source-external-modules
    resource: ../../internal/scaffold/external_modules.go
  - id: source-quickstart
    resource: ../../docs/content/quickstart.md
  - id: source-module
    resource: ../../internal/scaffold/module.go
  - id: source-module-publish
    resource: ../../internal/scaffold/module_publish.go
  - id: source-1
    resource: ../../cmd/dome/main.go
  - id: source-2
    resource: ../../internal/scaffold/scaffold.go
  - id: source-3
    resource: ../../internal/scaffold/publish.go
  - id: source-4
    resource: ../../internal/scaffold/publish_integration_test.go
  - id: source-5
    resource: ../../tests/integration/scaffold_test.go
---

# Project initialization

A dependency alone cannot install coding-agent skills or instructions. The initializer therefore embeds matching assets, validates the requested consumer module and canonical exact v0 release or resolved Go pseudo-version, resolves dependencies in a staged directory, and atomically publishes a fresh project. The sandbox and tracked placeholder keep the first exercise unprivileged and reproducible after cloning.

Consumer init requires `--project-module` to name the generated Go module,
without a compatibility alias for that identity. The identity need not
be hosted online. Resource-module init keeps its existing `--module` contract.
Repeatable consumer `--module path@version` inputs select external root Go module
packages exporting `Module`; generation assigns unique import aliases and emits
explicit `Module{}` registrations. Users configure backend or constructor
parameters in the generated Go before building. Staged dependency resolution
refuses changes to the requested Dome and external-module pins, preserving asset
alignment and avoiding publishing a project with silently changed dependencies.
The quickstart offers a pinned Go Docker shell with caller UID/GID and a bind
mount, keeping Go tooling in the container and project files in the working
directory. Keep agent invocations in the same mounted path for stable identity.

Consumer initialization from main uses Go's resolved immutable pseudo-version for
the dependency and every generated version field. Unversioned local builds and
mutable branch names cannot become dependency pins. Development guide links use
next; local Go documentation remains authoritative for the pinned API. Release
and pseudo-version consumer tests explicitly use a local replacement and do not
establish public main availability.

The standalone module initializer requires an exact release tag and operates in an existing empty directory and emits a public library, explicit unfinished backend, repository-local maintainer tooling and real container fixtures. It stages dependency/tool preparation, then uses no-replace entry publication. Partial published output is retained after failure to preserve concurrent writers. Test fixtures are not production host support.

## Boundary

The initializer supports Linux and fresh directories, not in-place merging into existing Go modules. Pre-release tests use an explicit local replacement until a real public tag exists.

Behavioral contract: [Project initialization](../../openspec/specs/project-initialization/spec.md).
