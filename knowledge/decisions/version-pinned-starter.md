---
type: Decision
title: "Project initialization"
description: "One release binds generated source, dependency, skill, and template metadata."
generated: {by: codex/2026-11-13, at: "2026-11-13T07:03:33Z"}
sources:
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

A dependency alone cannot install coding-agent skills or instructions. The initializer therefore embeds matching assets, validates the requested consumer module and exact release version, resolves dependencies in a staged directory, and atomically publishes a fresh project. The sandbox and tracked placeholder keep the first exercise unprivileged and reproducible after cloning.

## Boundary

The initializer supports Linux and fresh directories, not in-place merging into existing Go modules. Pre-release tests use an explicit local replacement until a real public tag exists.

Behavioral contract: [Project initialization](../../openspec/specs/project-initialization/spec.md).
