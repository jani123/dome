---
type: Decision
title: "Consumer authoring"
description: "Coding agents assist Go authoring while deployed agents remain independent of AI."
generated: {by: codex/2026-11-13, at: "2026-11-13T07:03:33Z"}
sources:
  - id: source-1
    resource: ../../internal/scaffold/templates/_agents/skills/dome/SKILL.md.tmpl
  - id: source-2
    resource: ../../internal/scaffold/templates/AGENTS.md.tmpl
  - id: source-3
    resource: ../../internal/scaffold/templates/README.md.tmpl
  - id: source-4
    resource: ../../internal/scaffold/scaffold_test.go
  - id: source-5
    resource: ../../docs/content/coding-agent.md
---

# Consumer authoring

The consumer skill is separate from maintainer rules. It compares the dependency and asset manifest before consulting installed APIs, generating resources, and checking plans. Apply remains tied to the user request. Upgrades compare a temporary version-matched project and preserve authored code. Maintainer OKF/OpenSpec tooling is not automatically injected into all generated consumer projects.

## Boundary

Packaging tests validate rendered metadata/guidance. Real Codex discovery and authoring need a live acceptance exercise and are not implied by static tests.

Behavioral contract: [Consumer authoring](../../openspec/specs/consumer-authoring/spec.md).
