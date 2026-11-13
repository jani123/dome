---
type: Operational Constraint
title: "Documentation publishing"
description: "Guides track release versions while pkg.go.dev supplies generated API reference."
generated: {by: codex/2026-11-13, at: "2026-11-13T07:03:33Z"}
sources:
  - id: source-1
    resource: ../../docs/hugo.toml
  - id: source-2
    resource: ../../docs/layouts/_default/baseof.html
  - id: source-3
    resource: ../../scripts/install-hugo.sh
  - id: source-4
    resource: ../../scripts/build-docs.sh
  - id: source-5
    resource: ../../scripts/build-site.sh
  - id: source-6
    resource: ../../.github/workflows/docs.yml
---

# Documentation publishing

Hugo renders project-owned Markdown/layouts under the GitHub project-site path. Version snapshots are rebuilt from their own immutable published tags and pinned tool versions. Development guides and the stable alias have separate meanings; prereleases cannot become latest stable. Public Pages availability still depends on repository settings and a successful deployed workflow.

## Boundary

Local HTML/link checks are not evidence of an externally deployed site. Owner Pages settings and deployment permissions need verification in GitHub.

Behavioral contract: [Documentation publishing](../../openspec/specs/documentation-publishing/spec.md).
