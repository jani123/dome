---
type: Operational Constraint
title: "Release distribution"
description: "A shared immutable tag keeps the library and authoring assets compatible."
generated: {by: codex/2026-11-13, at: "2026-11-13T07:03:33Z"}
sources:
  - id: source-1
    resource: ../../docs/content/releases.md
  - id: source-2
    resource: ../../.github/workflows/release.yml
  - id: source-3
    resource: ../../.github/workflows/ci.yml
  - id: source-4
    resource: ../../scripts/release.sh
  - id: source-5
    resource: ../../scripts/smoke-release.sh
---

# Release distribution

One repository version binds the library, initializer, templates, skill, examples and guide source. The first intended sequence is v0.1.0-rc.1 followed by v0.1.0. Native public-module smokes gate draft artifacts; cross-compilation is insufficient platform evidence. v0 patch releases preserve supported contracts, breaking changes require a minor and migration notes, and only the latest minor receives fixes initially.

## Boundary

Actual remote releases, package indexing, native arm64 runs and public bootstrap must be checked independently. A Git tag exposes module source before a GitHub Release is published.

Behavioral contract: [Release distribution](../../openspec/specs/release-distribution/spec.md).
