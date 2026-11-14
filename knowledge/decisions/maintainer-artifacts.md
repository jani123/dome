---
type: Decision
title: Maintainer artifact boundaries
description: Structured specs and curated knowledge replace committed ad hoc plans.
generated: {by: codex/2026-11-13, at: "2026-11-13T07:05:15Z"}
sources:
  - {id: rules, resource: ../../AGENTS.md}
  - {id: workflow, resource: ../../docs/development.md}
  - {id: validator, resource: ../../internal/knowledge/validate.go}
---

# Maintainer artifact boundaries

The owner rejected committed ad hoc plans and handovers because durable guidance
belongs in maintained documentation. Structured OpenSpec artifacts are the
explicit exception for proposals, design, tasks, verification and archives.
OpenSpec is reserved for features, architecture, contract or guarantee changes,
migrations, breaking upgrades, and significant cross-cutting design or acceptance.
Routine compatible dependency/security updates and maintenance proceed directly
with relevant verification and documentation. A patch restoring an existing
security guarantee does not itself change that guarantee. This boundary avoids
turning structured specifications into a maintenance log; actual impact and
explicit user requests determine when a proposal is warranted. Bug fixes retain
regression coverage; gaps or ambiguity in intended behavior also belong in the
affected feature requirements/scenarios, not solely in an executable test. This
does not require a new proposal for restoring an already specified contract.

Google OKF concepts preserve rationale and source-linked facts across sessions;
they must not become transcripts or renamed progress logs.

Repository rules, behavior specs, rationale, and usage guides serve different
readers. Keeping their authority explicit prevents a convenient memory note from
silently overriding the code, a requirement, or the user's direction. Agents read
only relevant indexes/concepts and follow evidence before changing behavior.

The format validator stays permissive where OKF is extensible. Dome's stricter
provenance and link checks are labeled policy rather than invented Google format
requirements. Automated checks cannot certify the truth of narrative claims.

See [maintainer workflow](../../docs/development.md) for commands and conventions.
