# Development environment

## Purpose

Retrospective specification of behavior inspected at `3e262920bbc20e41f843250f78a90c4935b9a036`.
This baseline documents existing implementation; it does not assert that these
requirements were authored before the code or that external acceptance passed.

## Requirements

### Requirement: Explicit egress policy

The devcontainer SHALL route allowed HTTP(S) hosts through Squid and document HTTPS Git use when direct DNS/SSH is unavailable.

Evidence: [.devcontainer/README.md](../../../.devcontainer/README.md).

#### Scenario: Explicit egress policy

- **WHEN** a contributor uses an SSH remote under the proxy-only network
- **THEN** the documentation identifies the transport mismatch rather than claiming github.com HTTPS is blocked

### Requirement: Real external-boundary verification

Required verification SHALL include unit/vet/format, race, real boundary integration, and compiled-agent container E2E through the documented Make targets.

Evidence: [AGENTS.md](../../../AGENTS.md).

#### Scenario: Real external-boundary verification

- **WHEN** a test touches actual filesystem, SQLite, keys, processes, or network
- **THEN** it is integration-tagged or part of the dedicated container suites

### Requirement: Owned Git transport fixture

Git integration SHALL use the project-owned image with standard packages, authenticated SSH and host verification, plus HTTP with explicit agent opt-in.

Evidence: [tests/integration/git_test.go](../../../tests/integration/git_test.go).

#### Scenario: Owned Git transport fixture

- **WHEN** Git integration runs
- **THEN** it exercises both real transports and authentication/known-host rejection without silently skipping or substituting a fake server

## Evidence and limitations

- [.devcontainer/README.md](../../../.devcontainer/README.md)
- [.devcontainer/proxy/allowed-domains.txt](../../../.devcontainer/proxy/allowed-domains.txt)
- [.devcontainer/compose.yaml](../../../.devcontainer/compose.yaml)
- [tests/integration/gitserver/Dockerfile](../../../tests/integration/gitserver/Dockerfile)
- [internal/testutil/containers.go](../../../internal/testutil/containers.go)
- [Makefile](../../../Makefile)
- [AGENTS.md](../../../AGENTS.md)

A privileged workspace is not a hardened isolation boundary. Live systemd coverage and external forge credentials are not required by the ordinary Git container fixture.

Rationale: [Proxy-only egress and real boundary fixtures keep network failures diagnosable.](../../../knowledge/operations/proxy-and-real-tests.md).
