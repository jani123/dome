# Release distribution

## Purpose

Retrospective specification of behavior inspected at `3e262920bbc20e41f843250f78a90c4935b9a036`.
This baseline documents existing implementation; it does not assert that these
requirements were authored before the code or that external acceptance passed.

## Requirements

### Requirement: Shared immutable release version

A single exact immutable tag SHALL identify library and initializer assets; published tags SHALL NOT be moved to repair a failed release.

Evidence: [docs/content/releases.md](../../../docs/content/releases.md).

#### Scenario: Shared immutable release version

- **WHEN** a published version is unusable
- **THEN** the documented repair is a new version and, when appropriate, retraction

### Requirement: Verified draft artifacts

A release tag SHALL run required checks, package Linux amd64/arm64 binaries plus source/checksums/notices, and require native public-tag smoke success before draft release creation.

Evidence: [.github/workflows/release.yml](../../../.github/workflows/release.yml).

#### Scenario: Verified draft artifacts

- **WHEN** one native platform smoke fails
- **THEN** the draft publication job does not run

### Requirement: Consumer-owned distribution

Dome release assets SHALL distribute the initializer/library materials rather than the demonstration agent or a universal customer agent.

Evidence: [scripts/release.sh](../../../scripts/release.sh).

#### Scenario: Consumer-owned distribution

- **WHEN** a customer deploys infrastructure definitions
- **THEN** they compile and distribute their own version-controlled agent

## Evidence and limitations

- [docs/content/releases.md](../../../docs/content/releases.md)
- [.github/workflows/release.yml](../../../.github/workflows/release.yml)
- [.github/workflows/ci.yml](../../../.github/workflows/ci.yml)
- [scripts/release.sh](../../../scripts/release.sh)
- [scripts/smoke-release.sh](../../../scripts/smoke-release.sh)

Actual remote releases, package indexing, native arm64 runs and public bootstrap must be checked independently. A Git tag exposes module source before a GitHub Release is published.

Rationale: [A shared immutable tag keeps the library and authoring assets compatible.](../../../knowledge/operations/release-and-compatibility.md).
