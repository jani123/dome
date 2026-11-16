# Proposal

## Why

Bare `agent run` currently reconciles once, duplicating `apply` and obscuring the supported continuous operation mode. Give `run` a clear foreground service role while retaining `apply` for one-shot execution.

## What Changes

- **BREAKING**: `agent run` reconciles immediately and repeats in the foreground with a default 30-second interval.
- Preserve positive `--interval` overrides; reject zero and negative intervals before reconciliation. Migrate one-shot callers to `agent apply`.
- Preserve sequential execution, cancellation, and exit-on-error supervisor handling.
- Update command help, operational guidance, and regression coverage.
- Exclude background daemonization, service installation, webhooks, automatic retries, new profile settings, and reconciliation/state changes.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `consumer-cli`: change [Supervised polling](../../specs/consumer-cli/spec.md) from opt-in polling to continuous foreground operation by default.

## Impact

Affects `cli/cli.go`, CLI tests and compiled-agent acceptance coverage, README, deployment documentation, and consumer operations knowledge. Scripts using bare `run` or `run --interval 0` must migrate to `apply`. Existing positive interval invocations remain compatible. No dependencies or persisted-state migrations are required.
