# Playbook execution

## Purpose

Retrospective specification of behavior inspected at `3e262920bbc20e41f843250f78a90c4935b9a036`.
This baseline documents existing implementation; it does not assert that these
requirements were authored before the code or that external acceptance passed.

## Requirements

### Requirement: Validated recovery contracts

Every step SHALL provide Observe, Apply, and Verify, and SHALL be repeatable or supply Recover; the graph SHALL reject invalid dependencies and cycles.

Evidence: [core/playbook/playbook_test.go](../../../core/playbook/playbook_test.go).

#### Scenario: Validated recovery contracts

- **WHEN** a non-repeatable step lacks recovery
- **THEN** validation rejects the workflow before execution

### Requirement: Dependency and lock scheduling

Independent ready nodes MAY run concurrently within the worker limit; dependencies and named locks SHALL constrain overlapping work.

Evidence: [core/dag/dag_test.go](../../../core/dag/dag_test.go).

#### Scenario: Dependency and lock scheduling

- **WHEN** one branch fails while an independent branch is ready
- **THEN** descendants of the failed branch remain blocked while independent work may finish

### Requirement: Observation-led resume

Completed nodes SHALL be re-observed and verified on resume; interrupted non-repeatable or uncertain work SHALL require appropriate recovery instead of blind apply.

Evidence: [core/dag/dag_test.go](../../../core/dag/dag_test.go).

#### Scenario: Observation-led resume

- **WHEN** a node completed externally before its success checkpoint
- **THEN** resume observes the host and establishes safe execution rather than trusting SQLite alone

### Requirement: Fingerprint and bounded retry

Resume SHALL reject changed workflow fingerprints; retries SHALL be bounded per invocation and callback deadlines SHALL use cancellable contexts.

Evidence: [core/dag/dag_test.go](../../../core/dag/dag_test.go).

#### Scenario: Fingerprint and bounded retry

- **WHEN** callback behavior changes for a resumable workflow
- **THEN** the author must bump Version and the changed fingerprint cannot resume the old execution

### Requirement: Failure branches preserve the original outcome

OnFailure recovery branches SHALL persist their own progress without hiding the
original failure or unblocking its dependent nodes.

Evidence: [DAG recovery tests](../../../core/dag/dag_test.go).

#### Scenario: Recovery branch succeeds

- **WHEN** a failed step's OnFailure branch completes successfully
- **THEN** the original failure is still returned and its descendants remain blocked

## Evidence and limitations

- [core/playbook/playbook.go](../../../core/playbook/playbook.go)
- [core/dag/dag.go](../../../core/dag/dag.go)
- [core/dag/dag_test.go](../../../core/dag/dag_test.go)
- [core/playbook/playbook_test.go](../../../core/playbook/playbook_test.go)

Callbacks must honor cancellation for timeouts to work. InProgress, Unknown, and Failed observations cannot justify blind repetition.

Rationale: [Explicit DAG edges and recovery contracts make interruptions safe to revisit.](../../../knowledge/architecture/dag-execution-and-recovery.md).
