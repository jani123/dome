# Workspace initialization verification

## Scope and provenance

Verified local working-tree implementation on 2026-11-14. This change adds
compiled-agent host initialization and saved runtime defaults; it does not
initialize a configuration repository or publish a release. Existing explicit
CLI/Go workflows remain supported. Runtime setup flags remain accepted for
compatibility and are visible only on init help.

The separately authorized routine security updates raise Go to 1.26.11, gRPC to
1.83.1 and the aligned OpenTelemetry API/SDK family to 1.45.0; both scaffold Go
minimums and their documentation match. Devbox Go 1.25 remains the documented
bootstrap for normal automatic compiler selection.

## Behavioral evidence

- Unit cases in cli/workspace_test.go verify interactive missing-only setup,
  complete unattended setup, non-terminal/JSON/EOF/cancellation failures, saved
  runtime defaults, source-unit switches, explicit false/empty flags, no backend
  calls during setup, invalid-profile failure, and setup/runtime help visibility.
  Reused help trees and the original cli.New function type are covered.
- workspace/profile_test.go covers strict schema, paths, defaults and rejected
  credential URLs. Real integration cases in workspace/store_integration_test.go
  cover profile permissions, round-trip/no-op/conflict behavior, symlinks,
  cancellation, concurrent writers, size bounds, documented YAML and cause
  preservation. workspace/publication_integration_test.go deterministically
  replaces the selected directory/ancestor and checks anonymous staging and
  preservation of concurrent files. Linux O_TMPFILE/hard-link/proc requirements
  are explicit in the public guide.
- Compiled-agent E2E initializes once, plans and applies twice, checks idempotency,
  selects the workspace from a different working directory, preserves operation
  history across repeated/conflicting setup, and rejects damaged profiles before
  changing managed files.
- The real project-owned Git HTTP/SSH suite initializes once and loads saved
  configuration, transport and encrypted-secret references, including after a
  container restart. Existing host-key, credential, retrieval, validation,
  decryption, transition and revision-preservation cases remain exercised.
- The generated consumer smoke initializes its local profile before plan/apply
  and retains stable sandbox ownership. Both scaffold minimums have assertions.

## Executed checks

The independent verification role ran task check, task test:race,
task test:e2e, task test:integration, task check:static, task check:security,
task check:artifacts, task docs:build and task check:workflow. All passed against
the final code. Check/race/E2E were rerun after the last constructor compatibility
edit. Real PTY exercises verified plain interactive setup, supplied-field prompt
omissions and init/plan/apply help. Isolated real-store probes verified successful
round-trip, oversized rejection without publication and preserved errno causes.

A standalone module was generated using actual RenderModule/PublishModule/
PrepareModule with an explicit local Dome replacement. Its setup, unit/format/vet,
race, real integration, container E2E, artifact, docs, workflow and security checks
passed. This is local source acceptance, not public tagged-module resolution or
native arm64 acceptance.

Core and generated module govulncheck scans report zero reachable vulnerabilities,
with one imported-package and three module-only advisories without observed
calls. No scans were suppressed. Scanner reachability is not a manual exploit
assessment or a permanent vulnerability-free guarantee.

Initial independent review identified and prompted corrections for repeated
supplied-field prompts, mismatched Save/Load size limits, discarded I/O causes,
directory/staging pathname races and HTTP userinfo credentials. Targeted regression
coverage and final checks above exercise those repairs. Formal role reports and
aggregate are ignored local review evidence rather than public approval.

Exact commands, compiler/tool versions, timestamps, exits, logs, PTY/store probes
and generated-module evidence are retained under ignored
`.build/notes/reviews/agent-workspace/final-checks/`. No required check was skipped.
No source commit, archive, remote publication or deployment is implied.
