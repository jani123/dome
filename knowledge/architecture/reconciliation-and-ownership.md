---
type: Architecture
title: "Reconciliation"
description: "Ownership and fresh verification prevent accidental pruning and false applied state."
generated: {by: codex/2026-11-13, at: "2026-11-13T07:03:33Z"}
sources:
  - id: source-1
    resource: ../../core/reconcile/reconcile.go
  - id: source-2
    resource: ../../core/reconcile/reconcile_test.go
  - id: source-3
    resource: ../../core/resource/resource.go
  - id: source-4
    resource: ../../tests/e2e/e2e_test.go
---

# Reconciliation

Persisted ownership restricts what may be pruned; it is not a substitute for observing the host. The runtime validates and observes the full plan before mutation, applies desired resources before pruning, records attempts before effects, and advances resource state only after verification. Earlier successful effects can remain after a later failure; there is no whole-host rollback.

## Boundary

Reconciliation is not an atomic host transaction. Changing configuration does not roll back earlier successful resources.

Behavioral contract: [Reconciliation](../../openspec/specs/reconciliation/spec.md).
