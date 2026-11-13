## ADDED Requirements

### Requirement: Structured change lifecycle

The repository SHALL use committed OpenSpec capability specs and change
artifacts for feature requirements, design, implementation tasks, verification,
and completed-change archives. Ad hoc plan and handover documents SHALL remain
outside version control.

#### Scenario: Plan a new feature

- **WHEN** a maintainer proposes behavior changes
- **THEN** the change identifies affected capabilities, observable requirements, design decisions, implementation work, and validation within the structured OpenSpec layout

#### Scenario: Preserve the working-notes rule

- **WHEN** an agent creates a standalone session handover outside OpenSpec
- **THEN** it remains local and lasting information is integrated into maintained documentation or curated knowledge

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
