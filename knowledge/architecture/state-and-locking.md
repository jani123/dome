---
type: Architecture
title: "Durable state"
description: "SQLite records ownership and attempts while host observations remain authoritative."
generated: {by: codex/2026-11-13, at: "2026-11-13T07:03:33Z"}
sources:
  - id: source-1
    resource: ../../core/state/state.go
  - id: source-2
    resource: ../../core/state/sqlite_integration_test.go
  - id: source-3
    resource: ../../core/dag/dag.go
  - id: source-4
    resource: ../../docs/content/recovery.md
---

# Durable state

Pure-Go SQLite avoids CGO deployment requirements and uses versioned migrations. A process guard serializes invocations sharing a database, while durable workflow records and named resource locks support recovery and concurrency. Different state databases must not own overlapping host resources. Backups require a stopped agent or a SQLite-aware backup method.

## Boundary

Persisted state proves neither current host state nor cross-database coordination. Binary downgrade after forward migrations is not promised.

Behavioral contract: [Durable state](../../openspec/specs/durable-state/spec.md).
