---
type: Architecture
title: "Resource transitions"
description: "A matching host target cannot bypass an unfinished linked workflow."
generated: {by: codex/2026-11-13, at: "2026-11-13T07:03:33Z"}
sources:
  - id: source-1
    resource: ../../core/reconcile/reconcile.go
  - id: source-2
    resource: ../../core/reconcile/reconcile_test.go
  - id: source-3
    resource: ../../examples/basic-agent/main.go
  - id: source-4
    resource: ../../tests/e2e/e2e_test.go
---

# Resource transitions

Resource transitions link an identity and non-secret target snapshot to a prepared workflow before effects. The persisted link survives a crash even when the host already matches the desired resource. Changing or removing that target is blocked until recovery finishes. Interrupted pruning must resume after the file has disappeared, then verify deletion before forgetting ownership.

## Boundary

Without a selected playbook, a transition implementation must explicitly declare its default apply safe, including deletion handles.

Behavioral contract: [Resource transitions](../../openspec/specs/resource-transitions/spec.md).
