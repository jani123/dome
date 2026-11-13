# Proposal

## Why

Dome's architecture, operating constraints, and completed work are currently spread
across source, tests, documentation, Git history, and conversations. Adopt curated
project memory and a consistent feature workflow so future sessions can recover
verified context and distinguish implemented behavior from unfinished work.

## What Changes

- Use Google's Open Knowledge Format (OKF v0.2) for a committed `knowledge/` bundle
  of source-linked facts, decisions, rationale, and operating constraints.
- Use OpenSpec for committed capability requirements and structured change
  proposals, designs, implementation tasks, verification, and archives.
- Keep the ban on ad hoc plans/handovers; explicitly permit OpenSpec artifacts and
  curated knowledge documentation.
- Import all existing capability areas from source, tests, README, documentation,
  and Git history. Capture completed behavior as a clearly labeled retrospective
  baseline, not invented historical proposals or completed checklists.
- Represent remaining onboarding/release acceptance as active work with explicit
  evidence requirements. Do not equate local implementation with public release.
- Add reproducible development-tool setup and CI validation, plus agent guidance
  for selective memory retrieval and keeping knowledge/specs current.

## Capabilities

### New Capabilities

- `project-knowledge`: Portable OKF project memory with provenance, progressive
  retrieval, ownership of facts, and validation.
- `spec-workflow`: OpenSpec change lifecycle, evidence-based baseline import,
  traceability, and maintenance alongside code and documentation.

### Modified Capabilities

None. No OpenSpec capability baseline exists yet; existing runtime behavior will
be documented and verified during migration without being redesigned.

## Impact

Affects `AGENTS.md`, `README.md`, developer documentation, `knowledge/`,
`openspec/`, generated maintainer skills, development tooling, and CI. The initial
OpenSpec change/configuration and rule exception establish this proposal; the
migration and ongoing tooling remain implementation tasks.

No new consumer runtime dependency, infrastructure DSL, hosted memory service,
Google Cloud/Gemini credentials, third-party OKF engine, or automatic rollout of
these maintainer tools into generated consumer projects is included.
