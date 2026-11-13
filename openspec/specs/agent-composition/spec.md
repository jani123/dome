# Agent composition

## Purpose

Retrospective specification of behavior inspected at `3e262920bbc20e41f843250f78a90c4935b9a036`.
This baseline documents existing implementation; it does not assert that these
requirements were authored before the code or that external acceptance passed.

## Requirements

### Requirement: Explicit module registration

Agent construction SHALL accept explicit modules and reject duplicate module names; core contracts SHALL remain independent of concrete modules.

Evidence: [core/registry/registry_test.go](../../../core/registry/registry_test.go).

#### Scenario: Explicit module registration

- **WHEN** a consumer registers the same module name twice
- **THEN** construction fails before reconciliation

### Requirement: External module composition

A separate Go module SHALL be able to implement and register resources through public Dome packages.

Evidence: [tests/integration/consumer_test.go](../../../tests/integration/consumer_test.go).

#### Scenario: External module composition

- **WHEN** an external consumer defines its own resource implementation
- **THEN** it can compile and produce a plan without importing internal packages

### Requirement: Configuration evaluation per invocation

Agent planning, reconciliation, and playbook execution SHALL retrieve and evaluate configuration for that invocation; CLI source flags SHALL override the configured default source.

Evidence: [agent/agent_integration_test.go](../../../agent/agent_integration_test.go).

#### Scenario: Configuration evaluation per invocation

- **WHEN** a consumer supplies a CLI configuration source
- **THEN** the invocation uses that source rather than the default

## Evidence and limitations

- [agent/agent.go](../../../agent/agent.go)
- [core/resource/resource.go](../../../core/resource/resource.go)
- [core/registry/registry.go](../../../core/registry/registry.go)
- [tests/integration/consumer_test.go](../../../tests/integration/consumer_test.go)
- [agent/example_test.go](../../../agent/example_test.go)

The demonstration agent contains intentional crash/failure hooks and is not the clean starter.

Rationale: [Explicit Go composition keeps the core independent of concrete modules.](../../../knowledge/architecture/library-composition.md).
