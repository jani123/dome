# Proposal

## Why

A single review pass can miss architecture, security, implementation defects, and plan drift, while claimed passing checks may not cover the reviewed files. One entrypoint should coordinate independent reviewers and actual verification against the same identified change.

## What Changes

- Add a repository-owned `dome-review` skill coordinating five separate specialist subagents: implementation analysis, architecture, security, verification runs, and OpenSpec alignment.
- Review the selected change and its staged, unstaged, and relevant untracked implementation files; identify scope precisely instead of reviewing every active proposal.
- Run formatting checks without rewriting files, linters/static analysis, unit/race tests, and required integration/E2E/artifact/docs checks. Record failed, blocked, and inapplicable checks distinctly.
- Add pinned static-analysis/security tooling and direct Task commands where current checks lack it; preserve one owner for running commands to avoid duplicate suites and shared-output races.
- Require every specialist to return a schema-validated structured report, including explicit no-findings and incomplete outcomes; produce a consolidated report using the same finding and evidence conventions.
- Merge evidence-backed findings into one report with severity, source locations, command outcomes, plan deviations, and remaining verification limits.
- Keep review read-only for source and OpenSpec task status. Do not auto-fix, commit, publish, archive, or rewrite a plan to hide drift. Temporary logs stay in ignored local storage.

## Capabilities

### New Capabilities

- `change-review`: Coordinated independent specialist review and reproducible verification of a scoped implementation.

### Modified Capabilities

None. Existing [spec-workflow](../../specs/spec-workflow/spec.md) and [development-environment](../../specs/development-environment/spec.md) define lifecycle and verification constraints; this capability owns the new coordinator without replacing upstream OpenSpec verification.

## Impact

New `.agents/skills/dome-review/` instructions, direct Task quality-check commands and pinned tool setup, maintainer README/development guidance and curated knowledge. This skill remains separate from the seven generated OpenSpec workflows. `use-taskfile-for-core-development` owns existing Make-wrapper migration; this change owns only new review-specific checks. Module-starter packaging is owned by `initialize-module-repository` and is not silently expanded here. No runtime library behavior changes.
