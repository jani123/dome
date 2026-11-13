---
type: Architecture
title: "Resource modules"
description: "Built-in resources expose injectable host boundaries and explicit supported states."
generated: {by: codex/2026-11-13, at: "2026-11-13T07:03:33Z"}
sources:
  - id: source-1
    resource: ../../modules/file/file.go
  - id: source-2
    resource: ../../modules/file/filesystem_integration_test.go
  - id: source-3
    resource: ../../modules/systemd/systemd.go
  - id: source-4
    resource: ../../modules/systemd/systemd_test.go
---

# Resource modules

The file module uses regular-file checks, explicit permissions, hashes, atomic replacement, and directory syncing. The systemd module manages existing loaded service state rather than installing services. Their boundaries are injectable so unit tests can reason about operations without mutating a host; real filesystem tests are separate integration checks.

## Boundary

Live systemd is not exercised by the normal containers because they do not run systemd. Service-unit installation and unsupported unit states are outside this module.

Behavioral contract: [Resource modules](../../openspec/specs/resource-modules/spec.md).
