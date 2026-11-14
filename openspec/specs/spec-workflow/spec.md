# Specification workflow

## Purpose

Maintain source-linked project context and verified feature changes through the
Google OKF and OpenSpec workflow introduced by
[adopt-project-knowledge-and-specs](../../changes/adopt-project-knowledge-and-specs/proposal.md).
Implementation evidence: [maintainer guide](../../../docs/development.md),
[validator](../../../internal/knowledge/validate.go),
[validator tests](../../../internal/knowledge/validate_test.go), and
[artifact CI](../../../.github/workflows/ci.yml).

## Requirements

### Requirement: Structured change lifecycle

The repository SHALL use committed OpenSpec specs and change artifacts for
features, architecture, public contracts, migrations, breaking upgrades, guarantee
changes, significant cross-cutting design or acceptance, and explicit user requests.
Routine compatible patches and maintenance SHALL proceed without new proposals
unless they meet those criteria. Ad hoc plans and handovers SHALL remain local.

#### Scenario: Plan a new feature

- **WHEN** a maintainer proposes a change meeting the OpenSpec scope criteria
- **THEN** the change identifies affected capabilities, observable requirements, design decisions, implementation work, and validation within the structured OpenSpec layout

#### Scenario: Routine dependency security patch

- **WHEN** a compatible dependency or toolchain patch preserves existing contracts and guarantees
- **THEN** maintainers implement and verify it directly without creating a proposal solely to record the update

#### Scenario: Maintenance reveals a migration

- **WHEN** an upgrade requires a migration, breaking contract change or architectural decision
- **THEN** the expanded scope is planned in OpenSpec before it is implemented

#### Scenario: Preserve the working-notes rule

- **WHEN** an agent creates a standalone session handover outside OpenSpec
- **THEN** it remains local and lasting information is integrated into maintained documentation or curated knowledge

### Requirement: Bug regression coverage and feature alignment

Bug fixes SHALL include appropriate executable regression coverage. When a bug
reveals missing, ambiguous or incorrect feature requirements, maintainers SHALL
update the affected OpenSpec requirements/scenarios. Restoring an already
specified contract SHALL NOT alone require a new proposal; changes to the intended
contract SHALL follow the OpenSpec scope criteria.

#### Scenario: Regression against specified behavior

- **WHEN** a bug violates an existing unambiguous requirement
- **THEN** the fix adds appropriate executable regression coverage without requiring a new proposal solely to record the fix

#### Scenario: Bug exposes a feature specification gap

- **WHEN** a bug reveals missing, ambiguous or incorrect intended behavior in a feature specification
- **THEN** the fix updates affected requirements/scenarios alongside executable regression coverage and uses a proposal if the intended contract changes

### Requirement: Evidence-based import of existing work

The initial migration SHALL cover all existing implementation and decision areas
identified in the migration inventory. Existing supported behavior SHALL be
captured as a retrospective capability baseline linked to implementation and test
evidence. The migration SHALL NOT fabricate historical proposals, implementation
dates, approvals, or completed tasks.

#### Scenario: Baseline crash recovery

- **WHEN** the migration imports already implemented transition recovery
- **THEN** the baseline states the recovery/ownership requirements and links their implementation/tests while identifying the specification as retrospectively imported

#### Scenario: Detect a coverage gap

- **WHEN** a documented capability lacks supporting implementation or acceptance evidence
- **THEN** the migration records that gap explicitly instead of certifying it as completed

### Requirement: Unfinished work remains active

Unfinished acceptance SHALL be represented as active OpenSpec work with explicit
completion evidence. A change SHALL be archived only after its scoped work is
actually implemented and verified; local implementation SHALL NOT substitute
for outstanding public delivery or platform validation.

#### Scenario: Release verification remains incomplete

- **WHEN** native arm64, live coding-agent, or public release/deployment checks remain unperformed
- **THEN** their acceptance work stays active even when local library tests have passed

### Requirement: Reproducible tools and coordinated maintenance

The repository SHALL provide pinned OpenSpec developer tooling and CI validation
without adding it to the consumer agent runtime. Changes SHALL update affected
capability specs, curated knowledge, and user documentation together according to
their distinct roles. Generated integration files SHALL preserve repository rules
and unrelated consumer assets.

#### Scenario: Initialize maintainer workflows

- **WHEN** the OpenSpec integration is installed or updated
- **THEN** it uses the selected version and workflow set without overwriting existing maintainer rules or the consumer Dome skill

#### Scenario: Validate a documentation-backed change

- **WHEN** a change alters a documented operating constraint or supported behavior
- **THEN** strict OpenSpec checks and knowledge-policy checks run alongside the relevant repository verification
