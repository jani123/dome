---
type: Decision
title: "Customer source access"
description: "GPLv3 source access supports customers changing their agents and service providers."
generated: {by: codex/2026-11-13, at: "2026-11-13T07:03:33Z"}
sources:
  - id: source-1
    resource: ../../LICENSE
  - id: source-2
    resource: ../../README.md
  - id: source-3
    resource: ../../docs/content/licensing.md
  - id: source-4
    resource: ../../docs/DEPENDENCY_REVIEW.md
  - id: source-5
    resource: ../../scripts/release.sh
---

# Customer source access

The owner accepts paid implementation and support but wants customers to retain their agent source and build materials. GPL-3.0-only without a linking exception was chosen for the distributed combined agent, not merely source changes to Dome. The deployment handoff also needs customer-controlled repositories, configuration, credentials and SOPS keys; source licensing alone cannot grant machine access.

## Boundary

This specification records project policy and documentation, not automated enforcement of downstream legal compliance. GPL permits compliant source-delivery methods other than physical bundling.

Behavioral contract: [Customer source access](../../openspec/specs/customer-source-access/spec.md).
