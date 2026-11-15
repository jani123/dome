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

The quickstart SHALL offer optional pinned-Go Docker bootstrap with a bind-mounted current directory and caller ownership. After project creation, it SHALL offer Dev Container CLI and VS Code alternatives for opening the generated development container. Subsequent build, test, plan, and apply steps SHALL run there without host Go or a hosted consumer module. Generated files SHALL remain host-visible.

#### Scenario: No host Go

- **WHEN** a Linux user follows Docker bootstrap and opens the generated devcontainer with the real CLI
- **THEN** build/test, plan, file creation and repeated no-op apply succeed in that workspace with caller-owned host-visible files

#### Scenario: Editor alternative

- **WHEN** a user selects VS Code
- **THEN** the guide explains opening the generated project with the Dev Containers extension and using Reopen in Container before running subsequent commands in its terminal

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

### Requirement: Consumer-owned development container

Consumer generation SHALL include a standalone Dev Container configuration and Dockerfile with pinned Go meeting the consumer minimum, a named non-root development user, caller UID alignment, and a stable bind-mounted workspace. It SHALL support Dev Container CLI and VS Code without host Go. Generation and container startup SHALL NOT apply resources or mount the Docker socket.

#### Scenario: Generated workspace

- **WHEN** a user generates a fresh consumer and starts its devcontainer
- **THEN** Go, build/test commands and agent commands run inside that container, project changes remain host-visible, and the development user can edit the project

#### Scenario: Re-enter workspace

- **WHEN** a user reopens the generated devcontainer
- **THEN** its project and state use the same workspace path and repeated apply reports no-op


## Evidence and limitations

- [cmd/dome/main.go](../../../cmd/dome/main.go)
- [internal/scaffold/scaffold.go](../../../internal/scaffold/scaffold.go)
- [internal/scaffold/publish.go](../../../internal/scaffold/publish.go)
- [internal/scaffold/publish_integration_test.go](../../../internal/scaffold/publish_integration_test.go)
- [tests/integration/scaffold_test.go](../../../tests/integration/scaffold_test.go)

The initializer supports Linux and fresh directories, not in-place merging into existing Go modules. Pre-release tests use an explicit local replacement until a real public tag exists.

Rationale: [One release binds generated source, dependency, skill, and template metadata.](../../../knowledge/decisions/version-pinned-starter.md).
