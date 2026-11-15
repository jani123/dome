## ADDED Requirements

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
