---
type: Operational Constraint
title: "Documentation publishing"
description: "Guides track release versions while local go doc describes the selected Go API."
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
  - id: source-preview
    resource: ../../scripts/serve-docs.sh
  - id: source-start
    resource: ../../.devcontainer/start-docs.sh
  - id: source-urls
    resource: ../../.devcontainer/urls.sh
  - id: source-6
    resource: ../../.github/workflows/docs.yml
---

# Documentation publishing

Hugo renders project-owned Markdown/layouts under the GitHub project-site path. Version snapshots are rebuilt from their own immutable published tags and pinned tool versions. Development guides and the stable alias have separate meanings; prereleases cannot become latest stable. Public Pages availability still depends on repository settings and a successful deployed workflow.

The devcontainer post-start hook starts Hugo automatically, waits for HTTP readiness, and uses a container-local process lock to avoid duplicate servers on repeated hooks. Startup logs live in ignored `.build/docs-preview.log`. Local preview uses pinned Hugo on container port 1313 with a dynamically assigned host-loopback binding. Host-side URL discovery reads current workspace mappings without starting services; mappings alone do not establish readiness. See the [preview guide](../../.devcontainer/README.md#local-documentation-preview).

## Boundary

Local HTML/link checks are not evidence of an externally deployed site. Owner Pages settings and deployment permissions need verification in GitHub.

Behavioral contract: [Documentation publishing](../../openspec/specs/documentation-publishing/spec.md).
