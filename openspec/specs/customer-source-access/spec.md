# Customer source access

## Purpose

Retrospective specification of behavior inspected at `3e262920bbc20e41f843250f78a90c4935b9a036`.
This baseline documents existing implementation; it does not assert that these
requirements were authored before the code or that external acceptance passed.

## Requirements

### Requirement: Declared library license

Dome SHALL declare GPL-3.0-only without a linking exception and document the combined-agent source-access purpose.

Evidence: [docs/content/licensing.md](../../../docs/content/licensing.md).

#### Scenario: Declared library license

- **WHEN** a reader evaluates the library license
- **THEN** the license and maintained guide explain the intended customer source/build access

### Requirement: Corresponding source distribution

Dome release packaging SHALL include the initializer source and vendored module source with licensing/notices alongside binary archives.

Evidence: [scripts/release.sh](../../../scripts/release.sh).

#### Scenario: Corresponding source distribution

- **WHEN** release packaging succeeds
- **THEN** a corresponding-source archive and applicable dependency notices are available

### Requirement: Customer control guidance

Documentation SHALL recommend handing source, dependency/build material and configuration to customers while privately transferring required credentials/keys; it SHALL distinguish managed applications from the covered combined agent.

Evidence: [docs/content/licensing.md](../../../docs/content/licensing.md).

#### Scenario: Customer control guidance

- **WHEN** a provider delivers a custom agent
- **THEN** the handoff guidance preserves customer rebuild/provider-choice capability without publishing secrets

## Evidence and limitations

- [LICENSE](../../../LICENSE)
- [README.md](../../../README.md)
- [docs/content/licensing.md](../../../docs/content/licensing.md)
- [docs/DEPENDENCY_REVIEW.md](../../../docs/DEPENDENCY_REVIEW.md)
- [scripts/release.sh](../../../scripts/release.sh)

This specification records project policy and documentation, not automated enforcement of downstream legal compliance. GPL permits compliant source-delivery methods other than physical bundling.

Rationale: [GPLv3 source access supports customers changing their agents and service providers.](../../../knowledge/decisions/customer-independence.md).
