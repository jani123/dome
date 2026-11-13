---
type: Architecture
title: "Playbook execution"
description: "Explicit DAG edges and recovery contracts make interruptions safe to revisit."
generated: {by: codex/2026-11-13, at: "2026-11-13T07:03:33Z"}
sources:
  - id: source-1
    resource: ../../core/playbook/playbook.go
  - id: source-2
    resource: ../../core/dag/dag.go
  - id: source-3
    resource: ../../core/dag/dag_test.go
  - id: source-4
    resource: ../../core/playbook/playbook_test.go
---

# Playbook execution

Step order in a slice is not a dependency contract. The scheduler follows graph edges and named locks, checkpoints attempts before effects, and re-observes completed and interrupted work. Callback authors must distinguish safe repetition from explicit recovery. Successful recovery does not conceal the original failure or unblock failed descendants.

## Boundary

Callbacks must honor cancellation for timeouts to work. InProgress, Unknown, and Failed observations cannot justify blind repetition.

Behavioral contract: [Playbook execution](../../openspec/specs/playbook-execution/spec.md).
