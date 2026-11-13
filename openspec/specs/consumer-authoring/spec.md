# Consumer authoring

## Purpose

Retrospective specification of behavior inspected at `3e262920bbc20e41f843250f78a90c4935b9a036`.
This baseline documents existing implementation; it does not assert that these
requirements were authored before the code or that external acceptance passed.

## Requirements

### Requirement: Pinned API authoring guidance

The generated Dome skill SHALL instruct agents to compare dependency, skill, and template versions and consult installed Go APIs rather than unconditionally using current-main documentation.

Evidence: [internal/scaffold/scaffold_test.go](../../../internal/scaffold/scaffold_test.go).

#### Scenario: Pinned API authoring guidance

- **WHEN** a dependency/manifest mismatch is detected
- **THEN** the guidance requires a reviewed synchronized update instead of silent API mixing

### Requirement: Authoring versus runtime

The starter SHALL use AI only as an optional authoring aid; the compiled agent SHALL operate without coding-agent service calls.

Evidence: [internal/scaffold/templates/cmd/agent/main.go.tmpl](../../../internal/scaffold/templates/cmd/agent/main.go.tmpl).

#### Scenario: Authoring versus runtime

- **WHEN** the customer runs their compiled agent
- **THEN** reconciliation uses the registered Go resources and configuration without an LLM dependency

### Requirement: User-directed apply and safe updates

Consumer instructions SHALL direct build/test and JSON plan inspection before requested apply and SHALL preserve custom source during manual version updates.

Evidence: [internal/scaffold/templates/README.md.tmpl](../../../internal/scaffold/templates/README.md.tmpl).

#### Scenario: User-directed apply and safe updates

- **WHEN** the user requests a version update
- **THEN** a temporary target-version project is reviewed and changes are merged without unattended replacement of authored resources

## Evidence and limitations

- [internal/scaffold/templates/_agents/skills/dome/SKILL.md.tmpl](../../../internal/scaffold/templates/_agents/skills/dome/SKILL.md.tmpl)
- [internal/scaffold/templates/AGENTS.md.tmpl](../../../internal/scaffold/templates/AGENTS.md.tmpl)
- [internal/scaffold/templates/README.md.tmpl](../../../internal/scaffold/templates/README.md.tmpl)
- [internal/scaffold/scaffold_test.go](../../../internal/scaffold/scaffold_test.go)
- [docs/content/coding-agent.md](../../../docs/content/coding-agent.md)

Packaging tests validate rendered metadata/guidance. Real Codex discovery and authoring need a live acceptance exercise and are not implied by static tests.

Rationale: [Coding agents assist Go authoring while deployed agents remain independent of AI.](../../../knowledge/decisions/ai-for-authoring-only.md).
