---
type: Operational Constraint
title: "Independent change review"
description: "Five specialist reviews share identified inputs and return validated evidence reports."
generated: {by: codex/2026-11-13, at: "2026-11-13T19:32:44.923537Z"}
sources:
  - id: source-skill
    resource: ../../.agents/skills/dome-review/SKILL.md
  - id: source-validator
    resource: ../../internal/review/report.go
  - id: source-checks
    resource: ../../Taskfile.yml
---

# Independent change review

A repository-owned skill coordinates separate implementation, architecture,
security, verification and plan-alignment agents. The command owner runs checks
once against identified inputs; formatting is inspected without source rewriting.
Versioned JSON reports preserve findings, source provenance, command outcomes and
limitations. A schema-valid report still requires evidence inspection; incomplete,
blocked or stale required work cannot establish a passing review.

See the [usage guide](../../docs/development.md#independent-change-review) and
[review capability](../../openspec/specs/change-review/spec.md).
Temporary transcripts/logs remain ignored rather than project knowledge.
