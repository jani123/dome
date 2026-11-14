# Proposal

## Why

Dome bootstraps consumer agents, but authors of reusable resource modules still assemble their own Go repository, specifications, knowledge, docs, and release tooling. A module initializer should establish these together in an empty directory so authors can start implementing host behavior with the public module contracts already represented.

## What Changes

- Add `dome module init --module <go-module-path> --name <resource-module-name>` operating in the current, existing empty directory.
- Generate a standalone Go library with public resource/module implementations, injectable host boundary, mock-based tests, and an external consumer registration example.
- Include runnable Ginkgo/Gomega and Testcontainers integration/E2E suites, a customizable Debian fixture, and a compiled test consumer demonstrating real container effects, idempotency, and pruning.
- Include pinned Hugo documentation build/serve tasks, a loopback-only dynamically assigned host port for the devcontainer docs server, and `task devcontainer:urls` to discover available service addresses.
- Include a project-adapted Devbox devcontainer with Docker-in-Docker, allowlisted proxy networking, isolated caches, and Task commands matching Dome development conventions.
- Include OpenSpec configuration, initial capability specs, pinned developer tooling and Codex workflows; an OKF v0.2 knowledge bundle; repository instructions; README and authoring/release guides.
- Use local `go doc` against the selected dependency or checkout as the API reference in generated documentation and authoring skills; do not request pkg.go.dev indexing or gate releases on its availability. Public module retrieval may still trigger third-party indexing.
- Generate `Taskfile.yml` as the sole task runner, with direct task implementations and consistent `task` commands in documentation and GitHub Actions; do not generate a Makefile or delegate to Make.
- Include build/test/artifact validation commands and GitHub CI plus tag-triggered, check-gated draft releases for a Go source library.
- Validate all inputs and exact Dome/template versions; refuse nonempty directories including hidden entries, resolve dependencies before publication, and preserve unrelated files on failures or races.
- Generate source locally only: no GitHub repository creation, Git initialization, tags, pushes, releases, resource application, or consumer-project modification.

## Capabilities

### New Capabilities

- `module-initialization`: Bootstrap and maintain the complete standalone resource-module repository starter, including safe empty-directory generation and its developer/release assets.

### Modified Capabilities

None. Existing [project-initialization](../../specs/project-initialization/spec.md) remains a fresh-directory consumer initializer. [consumer-authoring](../../specs/consumer-authoring/spec.md), [resource-modules](../../specs/resource-modules/spec.md), [project-knowledge](../../specs/project-knowledge/spec.md), [spec-workflow](../../specs/spec-workflow/spec.md), and [release-distribution](../../specs/release-distribution/spec.md) inform the new starter without extending their Dome-repository-specific requirements to third-party repositories.

## Impact

Changes will touch `cmd/dome`, `internal/scaffold`, new embedded module templates, generated-repository tests, and Dome authoring docs/knowledge. Public resource and reconciliation APIs remain unchanged. Generated repositories use public Dome APIs only; maintainer validators must be copied with appropriate notices or exposed through a deliberate public tooling boundary, never imported from Dome internal packages. OpenSpec and artifact tooling are developer dependencies only. This proposal adds neither an in-agent module installer nor a publishing service.
