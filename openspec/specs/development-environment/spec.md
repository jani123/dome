# Development environment

## Purpose

Retrospective specification of behavior inspected at `3e262920bbc20e41f843250f78a90c4935b9a036`.
This baseline documents existing implementation; it does not assert that these
requirements were authored before the code or that external acceptance passed.

Later Taskfile migration and docs-preview behavior are recorded in
[use-taskfile-for-core-development](../../changes/use-taskfile-for-core-development/proposal.md)
and [expose-core-docs-preview](../../changes/expose-core-docs-preview/proposal.md);
those additions are not attributed to the historical import revision.

## Requirements

### Requirement: Explicit egress policy

The devcontainer SHALL route allowed HTTP(S) hosts through Squid and document HTTPS Git use when direct DNS/SSH is unavailable.

Evidence: [.devcontainer/README.md](../../../.devcontainer/README.md).

#### Scenario: Explicit egress policy

- **WHEN** a contributor uses an SSH remote under the proxy-only network
- **THEN** the documentation identifies the transport mismatch rather than claiming github.com HTTPS is blocked

### Requirement: Real external-boundary verification

Required verification SHALL include unit/vet/format, race, real boundary integration, and compiled-agent container E2E through documented Taskfile tasks with no Make dependency.

Evidence: [AGENTS.md](../../../AGENTS.md).

#### Scenario: Real external-boundary verification

- **WHEN** a test touches actual filesystem, SQLite, keys, processes, or network
- **THEN** it is integration-tagged or part of the dedicated container suites

#### Scenario: Core checks without Make

- **WHEN** a contributor or CI invokes documented Taskfile commands with the supported tools installed
- **THEN** build, docs, tooling and required test checks run without delegating to Make

### Requirement: Owned Git transport fixture

Git integration SHALL use the project-owned image with standard packages, authenticated SSH and host verification, plus HTTP with explicit agent opt-in.

Evidence: [tests/integration/git_test.go](../../../tests/integration/git_test.go).

#### Scenario: Owned Git transport fixture

- **WHEN** Git integration runs
- **THEN** it exercises both real transports and authentication/known-host rejection without silently skipping or substituting a fake server

### Requirement: Discoverable local documentation preview

The devcontainer SHALL publish Hugo container port 1313 on a dynamically assigned host-loopback port. `task docs:serve` SHALL serve guides and `task devcontainer:urls` SHALL report current workspace bindings without starting containers or listing unrelated containers. Discovery SHALL distinguish mappings from readiness and explain remote Docker access.

#### Scenario: Discover the docs preview

- **WHEN** a contributor starts the devcontainer Hugo server and requests workspace URLs on the host
- **THEN** discovery reports the current mapped docs URL and guides are reachable there

#### Scenario: Start and restart the docs preview

- **WHEN** the devcontainer starts or restarts
- **THEN** its post-start hook starts Hugo in the background and reports success only after HTTP readiness, or fails with diagnostic log guidance

#### Scenario: Repeated startup hook

- **WHEN** the post-start hook runs again while its Hugo server is running
- **THEN** it reuses that server without starting a competing instance

#### Scenario: Workspace is stopped

- **WHEN** no matching workspace is running
- **THEN** discovery reports that state with startup guidance and does not start containers


#### Scenario: Concurrent workspaces

- **WHEN** two workspaces serve guides concurrently
- **THEN** each reports its own distinct dynamically assigned localhost URL

#### Scenario: Recreated workspace

- **WHEN** a workspace is recreated and URL discovery is run again
- **THEN** discovery reports current bindings without cached stale addresses

#### Scenario: Unrelated containers

- **WHEN** unrelated containers publish ports on the same Docker daemon
- **THEN** discovery excludes their bindings from the workspace results

## Evidence and limitations

- [.devcontainer/README.md](../../../.devcontainer/README.md)
- [.devcontainer/proxy/allowed-domains.txt](../../../.devcontainer/proxy/allowed-domains.txt)
- [.devcontainer/compose.yaml](../../../.devcontainer/compose.yaml)
- [tests/integration/gitserver/Dockerfile](../../../tests/integration/gitserver/Dockerfile)
- [internal/testutil/containers.go](../../../internal/testutil/containers.go)
- [Taskfile.yml](../../../Taskfile.yml)
- [AGENTS.md](../../../AGENTS.md)

A privileged workspace is not a hardened isolation boundary. Live systemd coverage and external forge credentials are not required by the ordinary Git container fixture.

Rationale: [Proxy-only egress and real boundary fixtures keep network failures diagnosable.](../../../knowledge/operations/proxy-and-real-tests.md).
