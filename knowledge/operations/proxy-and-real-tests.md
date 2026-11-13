---
type: Operational Constraint
title: "Development environment"
description: "Proxy-only egress and real boundary fixtures keep network failures diagnosable."
generated: {by: codex/2026-11-13, at: "2026-11-13T07:03:33Z"}
sources:
  - id: source-1
    resource: ../../.devcontainer/README.md
  - id: source-2
    resource: ../../.devcontainer/proxy/allowed-domains.txt
  - id: source-3
    resource: ../../.devcontainer/compose.yaml
  - id: source-4
    resource: ../../tests/integration/gitserver/Dockerfile
  - id: source-5
    resource: ../../internal/testutil/containers.go
  - id: source-6
    resource: ../../Makefile
  - id: source-7
    resource: ../../AGENTS.md
---

# Development environment

The devcontainer uses an internal network with blocked direct external DNS/traffic; Git over HTTPS uses the HTTP proxy while an SSH origin does not. Docker builds/runtimes need explicit compatible proxy routing. Real E2E and Git transport tests use isolated containers and the project-owned Debian Git/Apache/OpenSSH fixture; failures are not replaced by stubs or alternate downloads.

## Boundary

A privileged workspace is not a hardened isolation boundary. Live systemd coverage and external forge credentials are not required by the ordinary Git container fixture.

Behavioral contract: [Development environment](../../openspec/specs/development-environment/spec.md).
