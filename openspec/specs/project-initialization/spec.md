# Project initialization

## Purpose

Retrospective specification of behavior inspected at `3e262920bbc20e41f843250f78a90c4935b9a036`.
This baseline documents existing implementation; it does not assert that these
requirements were authored before the code or that external acceptance passed.

## Requirements

### Requirement: Version-matched generation

Consumer initialization SHALL require a valid external consumer module and a canonical v0 release version or canonical v0 Go pseudo-version resolved for the initializer. It SHALL pin the dependency, skill, templates, and manifest to the same immutable version, rejecting unversioned dev builds, mutable branch names, and unsupported version inputs.

Evidence: [internal/scaffold/scaffold_test.go](../../../internal/scaffold/scaffold_test.go).

#### Scenario: Version-matched generation

- **WHEN** a development initializer lacks explicit candidate metadata
- **THEN** init refuses to create an unpublished dependency pin

#### Scenario: Main bootstrap

- **WHEN** Go resolves an initializer invoked with `@main` to a canonical v0 pseudo-version
- **THEN** consumer initialization succeeds with that exact pseudo-version in its dependency and matching generated version metadata

#### Scenario: Release bootstrap

- **WHEN** the initializer is invoked at an exact supported release tag
- **THEN** the consumer dependency and generated metadata retain that exact release version

### Requirement: Safe project publication

Initialization SHALL refuse existing destinations and symlinked ancestors, resolve dependencies before atomic no-replace publication, and clean its own staged output after failure or cancellation.

Evidence: [internal/scaffold/publish_integration_test.go](../../../internal/scaffold/publish_integration_test.go).

#### Scenario: Safe project publication

- **WHEN** dependency resolution fails or a concurrent creator claims the destination
- **THEN** the initializer does not overwrite the destination or leave a published partial project

### Requirement: Clean sandbox consumer

The generated module SHALL contain typed configuration, consumer source/instructions, pinned metadata, the matching skill, and a tracked sandbox placeholder; generation SHALL NOT apply resources.

Evidence: [tests/integration/scaffold_test.go](../../../tests/integration/scaffold_test.go).

#### Scenario: Clean sandbox consumer

- **WHEN** the generated agent is built and plan is invoked
- **THEN** the sandbox file remains absent until explicit apply, then repeated apply reports no-op

### Requirement: Explicit consumer project module

Consumer initialization SHALL require the Go module identity of the project being created through `--project-module`, without requiring that identity to exist online. It SHALL preserve validation before publication; `--module` SHALL NOT act as an alias for project identity. Resource-module initialization SHALL retain its required `--module` argument.

#### Scenario: Project identity

- **WHEN** a user supplies a valid `--project-module` value
- **THEN** the generated Go module and consumer metadata use that value unchanged

#### Scenario: Missing or invalid identity

- **WHEN** the project identity is omitted or invalid
- **THEN** initialization fails before publication

### Requirement: Docker quickstart

The quickstart SHALL offer an optional Docker shell with pinned Go, a bind-mounted current directory, and the caller's UID/GID. The documented consumer bootstrap, build, test, plan, and apply SHALL run in that shell without host Go or a hosted consumer module, and generated files SHALL remain in the host working directory.

#### Scenario: No host Go

- **WHEN** a Linux user follows the Docker quickstart
- **THEN** the generated project and managed sandbox file are accessible in their working directory with caller ownership, and repeated apply reports no-op

### Requirement: External resource module selection

Consumer init SHALL accept repeatable `--module <module-path>@<exact-version>` inputs for root Go module packages exporting `Module`. It SHALL generate imports, explicit `Module{}` registrations and dependency pins. Users SHALL supply required registration parameters in generated Go. Invalid, duplicate, self, or built-in selections SHALL fail before publication; failed resolution or changed requested pins SHALL prevent publication.

#### Scenario: Register external modules

- **WHEN** valid pinned external modules are selected
- **THEN** the generated agent imports each root package using unique aliases and registers each Module explicitly alongside the file module

#### Scenario: Parameters require user configuration

- **WHEN** an external module needs backend or constructor parameters
- **THEN** initialization leaves an editable Module registration and explains the required manual configuration without applying resources

#### Scenario: Invalid or incompatible selection

- **WHEN** selections are invalid, duplicated, unresolvable, or resolve to different requested pins
- **THEN** initialization fails without publishing the consumer

## Evidence and limitations

- [cmd/dome/main.go](../../../cmd/dome/main.go)
- [internal/scaffold/scaffold.go](../../../internal/scaffold/scaffold.go)
- [internal/scaffold/publish.go](../../../internal/scaffold/publish.go)
- [internal/scaffold/publish_integration_test.go](../../../internal/scaffold/publish_integration_test.go)
- [tests/integration/scaffold_test.go](../../../tests/integration/scaffold_test.go)

The initializer supports Linux and fresh directories, not in-place merging into existing Go modules. Pre-release tests use an explicit local replacement until a real public tag exists.

Rationale: [One release binds generated source, dependency, skill, and template metadata.](../../../knowledge/decisions/version-pinned-starter.md).
