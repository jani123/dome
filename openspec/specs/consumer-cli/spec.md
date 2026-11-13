# Consumer CLI

## Purpose

Retrospective specification of behavior inspected at `3e262920bbc20e41f843250f78a90c4935b9a036`.
This baseline documents existing implementation; it does not assert that these
requirements were authored before the code or that external acceptance passed.

## Requirements

### Requirement: Inspectable commands

The consumer CLI SHALL expose run, plan, apply, status, modules, playbook listing/run/resume, operation listing/inspection, and version; plan and operation diagnostics SHALL support JSON.

Evidence: [cli/cli_test.go](../../../cli/cli_test.go).

#### Scenario: Inspectable commands

- **WHEN** the caller invokes plan --json
- **THEN** the CLI prints the plan without applying its described host resource changes

### Requirement: Public error categories

Exit codes SHALL be 0 for success, 2 for validation/configuration errors, and 1 for execution errors; oops boundaries SHALL preserve more specific categories and causes.

Evidence: [cli/cli_test.go](../../../cli/cli_test.go).

#### Scenario: Public error categories

- **WHEN** invalid CLI input reaches a validation boundary
- **THEN** the command reports the validation exit category

### Requirement: Supervised polling

Run SHALL reconcile once by default or poll at a positive requested interval, using injectable time; a reconciliation error SHALL end polling.

Evidence: [cli/timing_test.go](../../../cli/timing_test.go).

#### Scenario: Supervised polling

- **WHEN** a repeated run fails during reconciliation
- **THEN** the error is returned for supervisor handling rather than silently retried forever

### Requirement: Contextual categorized errors

Error boundaries SHALL use oops to preserve causes and specific failure
categories while attaching relevant non-secret module, resource, execution, and
node identifiers.

Evidence: [Error boundary tests](../../../core/fault/fault_test.go).

#### Scenario: A specific cause crosses a broader boundary

- **WHEN** an observation error is wrapped at an apply boundary
- **THEN** the specific category and original cause remain discoverable

## Evidence and limitations

- [cli/cli.go](../../../cli/cli.go)
- [cli/cli_test.go](../../../cli/cli_test.go)
- [cli/timing_test.go](../../../cli/timing_test.go)
- [agent/agent.go](../../../agent/agent.go)
- [core/fault/fault.go](../../../core/fault/fault.go)

Plan may open/create SQLite state even though it does not apply host resource changes. Polling is not a webhook service or automatic host rollback.

Rationale: [Plans and persisted operation diagnostics serve different purposes during recovery.](../../../knowledge/operations/consumer-operations.md).
