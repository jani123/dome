# Resource transitions

## Purpose

Retrospective specification of behavior inspected at `3e262920bbc20e41f843250f78a90c4935b9a036`.
This baseline documents existing implementation; it does not assert that these
requirements were authored before the code or that external acceptance passed.

## Requirements

### Requirement: Durable transition link

A selected transition SHALL use a registered validated playbook and checkpoint the prepared execution and target before workflow effects.

Evidence: [core/reconcile/reconcile_test.go](../../../core/reconcile/reconcile_test.go).

#### Scenario: Durable transition link

- **WHEN** a resource selects a transition workflow
- **THEN** the reconciler records the link before running that workflow

### Requirement: Pending target protection

An unfinished linked workflow SHALL resume even when the host matches its target, and changing/removing the pending target SHALL fail until recovery completes.

Evidence: [core/reconcile/reconcile_test.go](../../../core/reconcile/reconcile_test.go).

#### Scenario: Pending target protection

- **WHEN** configuration removes a resource with unfinished transition work
- **THEN** reconciliation reports recovery conflict instead of bypassing the pending workflow

### Requirement: Pruning recovery and verification

Interrupted pruning SHALL resume its workflow after the resource disappears; successful DAG completion SHALL NOT advance resource state without final host verification.

Evidence: [core/reconcile/reconcile_test.go](../../../core/reconcile/reconcile_test.go).

#### Scenario: Pruning recovery and verification

- **WHEN** a pruning workflow is interrupted after external deletion
- **THEN** resume completes the pending workflow and verifies deletion before forgetting ownership

### Requirement: Safe default transitions

A transition resource SHALL explicitly select a registered playbook or declare
its default apply safe, including for a restored pruning handle.

Evidence: [Transition contracts](../../../core/resource/resource.go) and
[transition tests](../../../core/reconcile/reconcile_test.go).

#### Scenario: No safe default or workflow is selected

- **WHEN** a transition resource returns no playbook and an unsafe default
- **THEN** planning fails instead of applying the change directly

## Evidence and limitations

- [core/reconcile/reconcile.go](../../../core/reconcile/reconcile.go)
- [core/reconcile/reconcile_test.go](../../../core/reconcile/reconcile_test.go)
- [examples/basic-agent/main.go](../../../examples/basic-agent/main.go)
- [tests/e2e/e2e_test.go](../../../tests/e2e/e2e_test.go)

Without a selected playbook, a transition implementation must explicitly declare its default apply safe, including deletion handles.

Rationale: [A matching host target cannot bypass an unfinished linked workflow.](../../../knowledge/architecture/transition-target-recovery.md).
