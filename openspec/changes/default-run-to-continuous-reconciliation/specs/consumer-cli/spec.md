## MODIFIED Requirements

### Requirement: Supervised polling

Run SHALL reconcile immediately and then sequentially repeat in the foreground, waiting 30 seconds after each successful reconciliation by default. A positive requested interval SHALL override that wait. Zero or negative intervals SHALL fail validation before reconciliation. Time SHALL be injectable; reconciliation errors SHALL end polling, and context cancellation SHALL interrupt waiting without another reconciliation. Apply SHALL remain the one-shot command.

#### Scenario: Supervised polling

- **WHEN** a repeated run fails during reconciliation
- **THEN** the error is returned for supervisor handling rather than silently retried forever

#### Scenario: Default foreground operation

- **WHEN** the caller invokes run without an interval
- **THEN** reconciliation begins immediately and repeats after a 30-second wait following each successful reconciliation until cancellation or failure
- **AND** the process remains in the foreground and reconciliations do not overlap

#### Scenario: Explicit polling interval

- **WHEN** the caller invokes run --interval 1m
- **THEN** it reconciles immediately and waits one minute after each successful reconciliation before starting the next

#### Scenario: Invalid polling interval

- **WHEN** the caller supplies a zero or negative interval
- **THEN** the CLI returns the validation error category without invoking reconciliation

#### Scenario: Cancellation while waiting

- **WHEN** the run context is canceled during the interval wait
- **THEN** waiting ends and the command returns without starting another reconciliation

#### Scenario: One-shot migration

- **WHEN** a caller uses apply for one-shot execution
- **THEN** the command reconciles once and exits without polling
