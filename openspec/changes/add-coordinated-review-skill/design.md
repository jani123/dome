# Design

## Context

The seven `.agents/skills/openspec-*` workflows are generated from pinned upstream OpenSpec and validated by `internal/knowledge/workflows.go`. Existing verify checks completeness/correctness/coherence but does not coordinate the requested independent roles. Taskfile currently provides Go formatting/vet/unit/race/integration/E2E and artifact checks, with some Make wrappers still present. CI separately runs actionlint, shell syntax, Hugo and dependency notices. No dedicated Staticcheck or govulncheck tasks currently exist.

## Goals / Non-Goals

**Goals:** one repository-owned review entrypoint, five independent perspectives, actual check execution and a review tied to identifiable source state.

**Non-Goals:** auto-fixing findings, replacing upstream OpenSpec workflows, changing archive rules, running exploit probes against real systems, or silently installing the skill into generated consumer/module repositories.

## Decisions

### One repository-owned entrypoint

Implement `.agents/skills/dome-review/SKILL.md` with optional concise role/report guidance in a referenced file. Keep it separate from generated OpenSpec assets so `specs:setup` cannot overwrite it. Use skill-creator validation and an independent behavioral exercise; no custom Python orchestration runtime is needed. Invoke as `$dome-review <change-name>` with an optional comparison revision; clarify ambiguous targets. Upstream OpenSpec verification remains available independently. The plan reviewer uses its completeness/correctness/coherence criteria without calling an artifact-mutating workflow.

### Scope and consistent inputs

Resolve OpenSpec context/status paths through the CLI. Read relevant AGENTS, knowledge, baseline specs, selected artifacts and Taskfile before dispatch. Record HEAD/base, diff and status including relevant untracked files, and content fingerprints for reviewed files in ignored `.build/notes/reviews/` storage. Give each specialist the same paths, artifacts and scope manifest rather than the coordinator's conclusions. Do not interpret all active proposals as missing requirements for one change. Whole-repository check results can be influenced by unrelated work; disclose that scope.

Review is read-only for source and task status. Before synthesizing findings compare reviewed file contents with the initial manifest, including creation/deletion. If they changed, rerun affected review/checks or mark results stale; avoid unbounded retries while another author is editing. Do not snapshot secrets into reports. Do not commit manifests, transcripts or logs.

### Five distinct reviewers within available slots

| Role | Responsibility |
| --- | --- |
| Implementation | Control flow, correctness, error causes/categories, cancellation, API behavior, regressions and test quality. |
| Architecture | Public module boundaries, explicit composition, configuration ownership, Observe/Decide/Apply/Verify/Persist, state/recovery/dependency invariants and maintainability. |
| Security | Input validation, paths/symlinks, subprocess boundaries, secret exposure, Git/SSH trust, dependency/workflow permissions and changed trust boundaries. |
| Verification | Execute the complete applicable command matrix once, checking formatting and real dependency/test outcomes. |
| Plan alignment | Map requirements/scenarios/design/tasks to source/tests; identify omissions, unintended scope, unsupported completion and plan contradictions. |

Spawn one separate specialist per role. With four total slots including coordinator, run at most three specialists at once and queue the rest; do not hardcode five-way parallel execution. Only the verification agent owns command runs with shared outputs; other reviewers inspect evidence and request additional checks through it. No need to fix a particular model override; inherit the active model. If delegation is unsupported, return an incomplete-review explanation rather than silently substituting coordinator self-review.

### Direct Task quality commands and verification selection

Add direct `task check:format`, `task check:static`, `task check:security` and `task check:workflow` tasks if absent. `check:format` checks gofmt and project gofumpt conventions without `-w`, using Go-aware file discovery including relevant untracked files and excluding ignored caches, vendor/generated artifacts only when explicitly appropriate. A quoted/path-safe implementation is needed; a tool error cannot look like an empty successful filename list. Preserve any repository requirements for generated Go templates through render/fixture tests rather than treating `.tmpl` as formatted Go source.

Pin Go-compatible Staticcheck and govulncheck releases during implementation and record the selected versions. Install tools in ignored `.build/tools` through a documented explicit setup task. Missing tools produce actionable failure, not a skipped pass. Use current repo actionlint and shell syntax checks for workflows/scripts; add no alternate downloads or automatic runtime fallback. Vulnerability findings must be contextualized with reachability evidence, without suppressing real failures.

Default Go implementation verification includes `task check` (format/vet/unit), static analysis and dependency security checks, with race and the complete required real integration/E2E suites for shared runtime or substantial feature changes under AGENTS. Artifact changes add `task check:artifacts`; docs changes add `task docs:build`; scripts/workflows add syntax/actionlint. Standalone formatting must still find changed Go files outside current hardcoded directory lists. Avoid rerunning an already covered suite when recorded output establishes the same source state. No auto-formatting or source-fixing options. Docs-only reviews may mark runtime suites inapplicable with a reason; user-required checks cannot be silently downgraded.

### Formal report contract

Ship a versioned JSON Schema and a deterministic report validator alongside the skill, referenced from its entrypoint. Every specialist returns one JSON report; narrative commentary may accompany it but cannot substitute for the structured result. Keep actual reports under ignored `.build/notes/reviews/<review-id>/`, never in committed ad hoc review documents. OpenSpec verification artifacts may preserve curated final evidence when appropriate.

Require these common fields:

- `schema_version`, `review_id`, `report_id`, `role`, and `input_id` matching the assigned review scope and content identity.
- `outcome`: `passed`, `findings`, `blocked`, `incomplete`, or `stale`.
- `coverage`: inspected paths, requirement/scenario identifiers where applicable, and performed versus unperformed checks with reasons.
- `findings`: an explicit array, empty when no findings; entries carry stable local ID, severity (`critical`, `high`, `medium`, `low`, `info`), title, path/line where applicable, triggering condition, impact, evidence references, suggested resolution, and whether the concern is demonstrated or still unverified.
- `evidence`: identified source references or command evidence; command records include exact invocation, working directory, tool versions, start/end time, exit code when executed, log reference, and outcome (`passed`, `failed`, `blocked`, `cancelled`, `not_applicable`) with a reason where needed.
- `limitations`: explicit array, including unavailable artifacts, environment constraints, or unperformed work; no secret contents in any report field.

The verification role must enumerate its selected command matrix, including commands that could not execute. Other roles can return an empty command list and source evidence. Role-specific schema constraints distinguish these cases. Reports claiming `passed` cannot contain unresolved actionable findings, blocked required checks, missing assigned coverage, or stale input identities. `not_applicable` requires an applicability reason and is not evidence of execution. Missing exit codes on unexecuted checks remain null; never invent an exit code or log.

The coordinator validates every report before synthesis. Request at most one correction for a malformed response, retaining the original as local diagnostic evidence. If still invalid or absent, record that role incomplete in the aggregate and continue supported analysis. Schema validation establishes structure, not correctness: the coordinator still checks claimed source/command evidence. Do not execute commands quoted inside a report merely to validate it.

Write `summary.json` with review/input identities, overall outcome, all five role statuses and report references, consolidated findings, command results, limitations and unresolved disagreements. Deduplicated findings include original `report_id`/finding IDs; retain specialist observations rather than overwrite raw reports. Aggregate schema validation and the Markdown summary must agree: a missing or invalid specialist, required failure/block, or stale state prevents `passed`. Do not average role scores into an apparent pass.

### Findings and final result

Each specialist returns status, findings (priority, path/line, condition, impact, evidence, suggested resolution), inspected scope and limits. Verification also returns command, tool version, elapsed time, exit outcome, and log location. Allow an explicit no-findings result; silence or a crashed agent is not a passing result. Root deduplicates by root cause, checks contradictions against source/evidence, and preserves specialist attribution. Report critical/high correctness/security defects before lower-severity issues. A missing plan or unreadable artifact is an evidence limitation, not automatically a fabricated requirement.

The final result contains scope, five-role coverage, actionable findings, checks and unresolved verification. Failed, blocked, cancelled, stale, or missing required work prevents a complete passing review. It does not auto-approve a merge, mark tasks complete, or assert deployment readiness.

## Risks / Trade-offs

- Five roles consume time and context → focused scope, bounded scheduling, one command owner and shared raw artifacts.
- Shared working tree changes → content identity checks and explicit stale-result reporting.
- New analysis tools may expose existing debt → distinguish changed-code findings from pre-existing failures while showing whole-check failure honestly; do not silently suppress them.
- Docker/network failures can obscure review → report endpoint/error evidence and retain blocked checks, never substitute fake integration.
- Skill prose may parse correctly yet behave poorly → independent forward-testing with deliberately flawed and clean disposable fixtures, not only front-matter validation.

## Migration Plan

Add the custom skill and new direct tasks; update maintainer docs and source-linked knowledge. Leave generated OpenSpec skills unchanged. `use-taskfile-for-core-development` remains responsible for migrating existing wrappers/CI. This change adds only new review checks and their documentation, and verifies refresh preserves the custom skill. Module initializer packaging remains separate future work unless explicitly added to its scope. Review adoption does not retroactively certify prior implementations.
