# Documentation publishing

## Purpose

Retrospective specification of behavior inspected at `3e262920bbc20e41f843250f78a90c4935b9a036`.
This baseline documents existing implementation; it does not assert that these
requirements were authored before the code or that external acceptance passed.

## Requirements

### Requirement: Pinned guide build

Guides SHALL build with pinned checksum-verified Hugo and checked internal paths under /dome/; API references SHALL link to pkg.go.dev at the guide version.

Evidence: [scripts/build-docs.sh](../../../scripts/build-docs.sh).

#### Scenario: Pinned guide build

- **WHEN** an RC guide is built
- **THEN** its bootstrap command and API reference use that RC version

### Requirement: Retained published snapshots

Published release guides SHALL remain under their version paths; next SHALL describe development and latest SHALL select a stable published release only.

Evidence: [scripts/build-site.sh](../../../scripts/build-site.sh).

#### Scenario: Retained published snapshots

- **WHEN** published versions include both stable and RC tags
- **THEN** site assembly retains both while latest selects the stable version

### Requirement: Trusted Pages deployment

Pages deployment SHALL use trusted main source, a successful build artifact, explicit deployment permissions, and the github-pages environment; PR checks SHALL not publish.

Evidence: [.github/workflows/docs.yml](../../../.github/workflows/docs.yml).

#### Scenario: Trusted Pages deployment

- **WHEN** the documentation workflow deploys
- **THEN** its deployment depends on the completed build and uses the configured Pages environment

## Evidence and limitations

- [docs/hugo.toml](../../../docs/hugo.toml)
- [docs/layouts/_default/baseof.html](../../../docs/layouts/_default/baseof.html)
- [scripts/install-hugo.sh](../../../scripts/install-hugo.sh)
- [scripts/build-docs.sh](../../../scripts/build-docs.sh)
- [scripts/build-site.sh](../../../scripts/build-site.sh)
- [.github/workflows/docs.yml](../../../.github/workflows/docs.yml)

Local HTML/link checks are not evidence of an externally deployed site. Owner Pages settings and deployment permissions need verification in GitHub.

Rationale: [Guides track release versions while pkg.go.dev supplies generated API reference.](../../../knowledge/operations/versioned-documentation.md).
