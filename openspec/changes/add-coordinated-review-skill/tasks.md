# Tasks

## 1. Skill coordinator and specialist contracts

- [x] 1.1 Create repository-owned `dome-review` entrypoint with scoped invocation, five distinct reviewer prompts, bounded scheduling, and no-source-write boundaries; validate with skill-creator tooling and check that upstream OpenSpec generation preserves it.
- [x] 1.2 Add input-state identification, reviewer result expectations, stale-source handling and final evidence/status aggregation; verify fixtures cover mixed working trees, untracked source, duplicate findings, a missing specialist, and changed inputs without claiming a passing review.
- [x] 1.3 Add versioned JSON Schemas for common/role-specific reports and the aggregate plus a deterministic validator; verify valid empty-findings, findings and blocked reports, rejection of missing fields/invalid roles/outcomes, identity mismatches, and forbidden passing outcomes with failed required checks.
- [x] 1.4 Document invocation, specialist coverage, report interpretation formal report fields, aggregate provenance, correction handling, and local-log policy in maintainer README/development guidance and source-linked knowledge; verify links and artifact checks.

## 2. Reproducible verification commands

- [x] 2.1 Add direct non-mutating formatting checks covering gofmt/gofumpt and relevant new Go files; verify clean and unformatted disposable files including paths with spaces, command errors, and a file outside existing hardcoded package directories, with unchanged file contents after checking.
- [x] 2.2 Pin Go-compatible Staticcheck and govulncheck tools and add direct setup/static/security Task commands with documented prerequisites; verify actual tool versions, successful execution where supported, and clear nonzero outcomes for missing tools or real dependency/network failures.
- [x] 2.3 Add a direct workflow/script check task reusing pinned actionlint and shell syntax conventions; verify representative malformed fixtures fail and valid project workflows/scripts pass, without fixing input files.
- [x] 2.4 Implement verification-agent check selection and single execution ownership using current AGENTS/Task requirements; verify a Go runtime scope requires full applicable suites and a docs-only scope records reasoned applicability without silently dropping requested checks.

## 3. Independent behavioral acceptance

- [x] 3.1 Forward-test the skill through an independent coordinator in an isolated disposable repository with a realistic implementation and plan; verify five separate specialist agents return schema-valid reports within available slots, the aggregate preserves their identities/provenance, and real commands run once under the verification owner.
- [x] 3.2 Exercise a flawed fixture containing an implementation defect, architecture boundary violation, security issue, plan mismatch and unformatted new Go file; verify evidence-backed findings, unchanged source/task state, and an explicitly non-passing result. Do not give evaluating agents the expected findings.
- [x] 3.3 Exercise a clean fixture, unavailable Docker/tooling, interrupted specialist and changed input state; verify honest passed/blocked/incomplete/stale reporting, malformed-report correction and failure handling, aggregate/Markdown consistency, no fabricated findings or passes, and no unrelated source mutation.
- [ ] 3.4 Run the applicable project Go/tooling, workflow, docs and artifact checks for the delivered skill/check changes; preserve actual logs in ignored storage and record only durable scoped verification evidence in OpenSpec.

## Workflow follow-up

- Verify against capability scenarios, sync the new spec, and archive only completed verified scope.
- Core Make-wrapper migration stays in `use-taskfile-for-core-development`; generated module packaging stays in `initialize-module-repository` unless separately approved.
