# Reconciliation

## Purpose

Retrospective specification of behavior inspected at `3e262920bbc20e41f843250f78a90c4935b9a036`.
This baseline documents existing implementation; it does not assert that these
requirements were authored before the code or that external acceptance passed.

## Requirements

### Requirement: Observation before mutation

Reconciliation SHALL validate desired resources and observe the planned changes before mutation, following Observe → Decide → Apply → Verify → Persist.

Evidence: [core/reconcile/reconcile_test.go](../../../core/reconcile/reconcile_test.go).

#### Scenario: Observation before mutation

- **WHEN** a desired resource fails validation or observation
- **THEN** the invocation fails before applying or pruning resources

### Requirement: Owned pruning only

Pruning SHALL remove only previously managed identities and SHALL restore their validated deletion handles through registered modules.

Evidence: [core/reconcile/reconcile_test.go](../../../core/reconcile/reconcile_test.go).

#### Scenario: Owned pruning only

- **WHEN** desired state omits an unowned host file
- **THEN** the unrelated file remains untouched

### Requirement: Verified applied state

The reconciler SHALL persist an attempt before external effects and SHALL update applied ownership/snapshots only after fresh observation matches the target.

Evidence: [core/reconcile/reconcile_test.go](../../../core/reconcile/reconcile_test.go).

#### Scenario: Verified applied state

- **WHEN** an apply operation fails verification
- **THEN** the failed resource does not advance its applied state

### Requirement: Overall revision on success

The overall applied Git revision SHALL advance only after all reconciliation work succeeds; earlier verified resource effects MAY remain after later failure.

Evidence: [core/reconcile/reconcile_test.go](../../../core/reconcile/reconcile_test.go).

#### Scenario: Overall revision on success

- **WHEN** a later resource fails
- **THEN** the overall revision remains at its last successful value

## Evidence and limitations

- [core/reconcile/reconcile.go](../../../core/reconcile/reconcile.go)
- [core/reconcile/reconcile_test.go](../../../core/reconcile/reconcile_test.go)
- [core/resource/resource.go](../../../core/resource/resource.go)
- [tests/e2e/e2e_test.go](../../../tests/e2e/e2e_test.go)

Reconciliation is not an atomic host transaction. Changing configuration does not roll back earlier successful resources.

Rationale: [Ownership and fresh verification prevent accidental pruning and false applied state.](../../../knowledge/architecture/reconciliation-and-ownership.md).
