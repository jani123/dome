---
type: Operational Constraint
title: "Development environment"
description: "Proxy-only egress and real boundary fixtures keep network failures diagnosable."
generated: {by: codex/2026-11-13, at: "2026-11-13T07:03:33Z"}
sources:
  - id: module
    resource: ../../go.mod
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
    resource: ../../Taskfile.yml
  - id: source-7
    resource: ../../AGENTS.md
---

# Development environment

The devcontainer uses an internal network with blocked direct external DNS/traffic; Git over HTTPS uses the HTTP proxy while an SSH origin does not. Docker builds/runtimes need explicit compatible proxy routing. Real E2E and Git transport tests use isolated containers and the project-owned Debian Git/Apache/OpenSSH fixture; failures are not replaced by stubs or alternate downloads.

Nested developer environments explicitly chain their proxy through the surrounding allowed proxy when direct egress is unavailable. Each proxy needs a distinct visible hostname; reusing the parent identity can reject chained HTTP package requests. Generated module proxies use the module name, while both layers keep their allowlists and destination checks. The [network guide](../../.devcontainer/README.md#running-inside-another-restricted-devcontainer) documents the explicit customization rather than an automatic download fallback.

Development uses the compiler minimum declared in go.mod (Go 1.26.11 or later).
Devbox supplies Go 1.25 as a bootstrap; automatic toolchain selection downloads
the required patched compiler from the normal Go distribution endpoints. Both
repository initializer templates declare the same compiler minimum, and CI reads
go.mod. Disabling automatic selection requires providing a sufficiently recent
compiler explicitly. Fixed gRPC and aligned OpenTelemetry dependencies are pinned
in go.mod; a successful historical scan is not a permanent security guarantee.

Generated-module smoke checks build and test an external module directly. Full
native acceptance instead uses each documented Taskfile check once, including
race and real container suites; it does not repeat standalone validators or
compile-only container probes. Named command subtests expose timings in verbose
CI output. Native acceptance uses a separate setup-go cache key from lightweight
jobs so their smaller caches cannot win publication of the same cache entry.

## Boundary

A privileged workspace is not a hardened isolation boundary. Live systemd coverage and external forge credentials are not required by the ordinary Git container fixture.

Behavioral contract: [Development environment](../../openspec/specs/development-environment/spec.md).
