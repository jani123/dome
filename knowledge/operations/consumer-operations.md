---
type: Operational Constraint
title: "Consumer CLI"
description: "Plans and persisted operation diagnostics serve different purposes during recovery."
generated: {by: codex/2026-11-13, at: "2026-11-13T07:03:33Z"}
sources:
  - id: source-1
    resource: ../../cli/cli.go
  - id: source-2
    resource: ../../cli/cli_test.go
  - id: source-3
    resource: ../../cli/timing_test.go
  - id: source-4
    resource: ../../agent/agent.go
  - id: source-5
    resource: ../../core/fault/fault.go
---

# Consumer CLI

The compiled CLI can observe plans, reconcile, inspect durable execution history, and run or resume workflows without an AI runtime. A JSON plan reflects a fresh observation; persisted status is historical evidence, not a host probe. Polling exits on reconciliation failure so a supervisor can report and restart it. Error boundaries preserve specific causes/categories without adding secrets.

## Boundary

Plan may open/create SQLite state even though it does not apply host resource changes. Polling is not a webhook service or automatic host rollback.

Behavioral contract: [Consumer CLI](../../openspec/specs/consumer-cli/spec.md).
