# Documentation publishing

## Purpose

Retrospective specification of behavior inspected at `3e262920bbc20e41f843250f78a90c4935b9a036`.
This baseline documents existing implementation; it does not assert that these
requirements were authored before the code or that external acceptance passed.

Local API guidance was subsequently revised by
[use-local-api-documentation](../../changes/use-local-api-documentation/proposal.md);
that revision is not attributed to the historical import above.

## Requirements

### Requirement: Pinned guide build

Guides SHALL build with pinned checksum-verified Hugo and checked internal paths under /dome/; API reference guidance SHALL use local `go doc` against the selected checkout or dependency version, without requiring pkg.go.dev indexing or availability.

Evidence: [scripts/build-docs.sh](../../../scripts/build-docs.sh).

#### Scenario: Pinned guide build

- **WHEN** an RC guide is built
- **THEN** its bootstrap command uses that RC version and API guidance explains selecting that version locally before running `go doc`

#### Scenario: Hosted API index unavailable

- **WHEN** pkg.go.dev is unavailable or has not indexed the selected version
- **THEN** users can inspect that version with local `go doc` after its source is available and no release gate waits for hosted documentation

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

Rationale: [Guides track release versions while local go doc describes the selected Go API.](../../../knowledge/operations/versioned-documentation.md).
