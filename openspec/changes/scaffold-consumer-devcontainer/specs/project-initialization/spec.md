## ADDED Requirements

### Requirement: Consumer-owned development container

Consumer generation SHALL include a standalone Dev Container configuration and Dockerfile with pinned Go meeting the consumer minimum, a named non-root development user, caller UID alignment, and a stable bind-mounted workspace. It SHALL support Dev Container CLI and VS Code without host Go. Generation and container startup SHALL NOT apply resources or mount the Docker socket.

#### Scenario: Generated workspace

- **WHEN** a user generates a fresh consumer and starts its devcontainer
- **THEN** Go, build/test commands and agent commands run inside that container, project changes remain host-visible, and the development user can edit the project

#### Scenario: Re-enter workspace

- **WHEN** a user reopens the generated devcontainer
- **THEN** its project and state use the same workspace path and repeated apply reports no-op

## MODIFIED Requirements

### Requirement: Docker quickstart

The quickstart SHALL offer optional pinned-Go Docker bootstrap with a bind-mounted current directory and caller ownership. After project creation, it SHALL offer Dev Container CLI and VS Code alternatives for opening the generated development container. Subsequent build, test, plan, and apply steps SHALL run there without host Go or a hosted consumer module. Generated files SHALL remain host-visible.

#### Scenario: No host Go

- **WHEN** a Linux user follows Docker bootstrap and opens the generated devcontainer with the real CLI
- **THEN** build/test, plan, file creation and repeated no-op apply succeed in that workspace with caller-owned host-visible files

#### Scenario: Editor alternative

- **WHEN** a user selects VS Code
- **THEN** the guide explains opening the generated project with the Dev Containers extension and using Reopen in Container before running subsequent commands in its terminal
