## MODIFIED Requirements

### Requirement: Real external-boundary verification

Required verification SHALL include unit/vet/format, race, real boundary integration, and compiled-agent container E2E through documented Taskfile tasks with no Make dependency.

Evidence: [AGENTS.md](../../../AGENTS.md).

#### Scenario: Real external-boundary verification

- **WHEN** a test touches actual filesystem, SQLite, keys, processes, or network
- **THEN** it is integration-tagged or part of the dedicated container suites

#### Scenario: Core checks without Make

- **WHEN** a contributor or CI invokes documented Taskfile commands with the supported tools installed
- **THEN** build, docs, tooling and required test checks run without delegating to Make
