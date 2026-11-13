# Durable state

## Purpose

Retrospective specification of behavior inspected at `3e262920bbc20e41f843250f78a90c4935b9a036`.
This baseline documents existing implementation; it does not assert that these
requirements were authored before the code or that external acceptance passed.

## Requirements

### Requirement: Pure Go durable storage

State SHALL use database/sql and modernc.org/sqlite with transactional versioned migrations and separate records for ownership, applied snapshots, and execution attempts.

Evidence: [core/state/sqlite_integration_test.go](../../../core/state/sqlite_integration_test.go).

#### Scenario: Pure Go durable storage

- **WHEN** an existing version-one database is opened
- **THEN** supported migrations preserve its data and expose attempt history

### Requirement: Invocation guard

Agent invocations sharing a state database SHALL hold the process guard through operations; direct store/DAG callers SHALL hold that guard themselves.

Evidence: [core/state/sqlite_integration_test.go](../../../core/state/sqlite_integration_test.go).

#### Scenario: Invocation guard

- **WHEN** another invocation already holds the same database guard
- **THEN** the contender does not run concurrent reconciliation against that store

### Requirement: Atomic named locks

Resource lock acquisition SHALL acquire all requested names atomically or report contention without partial ownership.

Evidence: [core/state/sqlite_integration_test.go](../../../core/state/sqlite_integration_test.go).

#### Scenario: Atomic named locks

- **WHEN** one requested lock is held by another execution
- **THEN** acquisition fails without retaining the other requested locks

### Requirement: Non-secret persistence

Snapshots and execution metadata SHALL contain non-secret representations, not configuration blobs, plaintext secrets, or authentication material.

Evidence: [modules/file/file_test.go](../../../modules/file/file_test.go).

#### Scenario: Non-secret persistence

- **WHEN** a file resource is persisted
- **THEN** its snapshot contains a content hash and permissions rather than file contents

## Evidence and limitations

- [core/state/state.go](../../../core/state/state.go)
- [core/state/sqlite_integration_test.go](../../../core/state/sqlite_integration_test.go)
- [core/dag/dag.go](../../../core/dag/dag.go)
- [docs/content/recovery.md](../../../docs/content/recovery.md)

Persisted state proves neither current host state nor cross-database coordination. Binary downgrade after forward migrations is not promised.

Rationale: [SQLite records ownership and attempts while host observations remain authoritative.](../../../knowledge/architecture/state-and-locking.md).
