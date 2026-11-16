# Design

## Context

See proposal.md for motivation. `cli/cli.go` already runs a sequential Apply/showPlan/Sleep loop through an injectable clock; zero currently exits after one iteration. `cli/timing_test.go` covers an explicit interval through fake time. README documents the old default, and deployment guidance requires an explicit interval. The existing consumer-cli requirement explicitly specifies one-shot default behavior and must change.

## Goals / Non-Goals

**Goals:** Make foreground polling the default, retain existing failure and cancellation propagation, and provide an explicit migration path.

**Non-Goals:** Add supervisor configuration, background processes, profile schema fields, or modify the reconciliation engine and recovery guarantees.

## Decisions

- Set the CLI duration default to 30 seconds and validate strictly positive values before backend Apply. Remove the zero-duration one-shot branch. Keeping zero as an alias was considered, but `apply` already provides that behavior and a positive-only contract makes run unambiguous.
- Keep the existing fixed delay after completion, rather than fixed-rate ticking. This preserves sequential execution and avoids accumulated ticks or overlapping reconciliation when work exceeds the interval.
- Reuse injectable time and command context. Preserve existing cancellation/error propagation and output handling; do not add retries or change exit categories.
- Keep the default in the consumer CLI, with no persisted setting. Main specs define the contract; delta artifacts describe the proposal until sync/archive. README and deployment docs explain operation/migration; curated knowledge preserves the command-role rationale with source references. Update help consistently.

## Risks / Trade-offs

- Existing scripts expecting bare run to exit may hang → prominently document migration to apply and test the distinction.
- Explicit zero intervals become validation failures → test zero and negative inputs before backend invocation and document migration.
- Polling increases repeated host work for bare run users → retain the 30-second delay and configurable positive override; preserve exit-on-error behavior.

## Migration Plan

Consumers rebuild their agents to receive the changed default. Replace one-shot `run` and `run --interval 0` invocations with `apply`; leave positive interval invocations unchanged. No database or runtime profile migration is needed. Reverting the CLI change restores the old default without state conversion. Verification remains pending until implementation and the listed checks run.
