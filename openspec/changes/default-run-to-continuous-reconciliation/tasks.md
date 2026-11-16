# Tasks

## 1. Continuous foreground command contract

- [x] 1.1 Change `cli/cli.go` to default run to 30 seconds, reject non-positive intervals before Apply, remove one-shot run behavior, and update help; verify help shows the default and positive override.
- [x] 1.2 Extend injected-clock CLI regression coverage for immediate first execution, default and overridden waits, sequential calls, zero/negative validation without Apply, cancellation without another call, reconciliation failure without retry, and unchanged one-shot apply; run the relevant CLI unit tests without real sleeps.
- [x] 1.3 Update README, `docs/content/deployment.md`, and `knowledge/operations/consumer-operations.md` with foreground operation, interval semantics, supervisor failure behavior, and migration from bare/zero-interval run to apply. Preserve provenance and check other source/generated instructions for old claims; verify help and examples agree and render changed deployment guidance using the documented docs checks.

## 2. Acceptance and repository verification

- [x] 2.1 Add compiled-agent container acceptance coverage for bare run remaining active through repeated reconciliation, an observable configuration update, and termination; retain apply one-shot coverage. Exercise real effects with isolated storage and verify the focused E2E scenario passes.
- [x] 2.2 Run `task check:artifacts` and strict OpenSpec validation; verify maintained artifact links and delta consistency without claiming pending acceptance passed.
- [x] 2.3 Run `task check`, `task test:race`, `task test:e2e`, and `task test:integration`; record actual results and any failing or unavailable checks before marking verification complete.
