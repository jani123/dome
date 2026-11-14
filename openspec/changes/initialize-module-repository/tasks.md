# Tasks

## 1. Command and render contract

- [x] 1.1 Add module-init command inputs, exact version validation, and pure module rendering; verify mocked command tests cover required flags, invalid paths/names/versions, help, exit categories, and unchanged consumer init.
- [x] 1.2 Add the module template inventory, provenance/license metadata, and executable-file modes; verify render tests check formatted Go, exact version agreement, safe relative paths, expected files, and no consumer sandbox assets.
- [x] 1.3 Document invocation, empty-directory prerequisite including `.git`, and supported platforms in Dome README/authoring docs; verify examples match CLI help and update source-linked curated knowledge.

## 2. Resource library starter

- [x] 2.1 Implement templates for public Module, Resource, Backend, categorized unfinished adapter, validated restore and safe-default transition; verify generated Testify tests cover identity, restoration, snapshots, all actions, cancellation/error propagation, and unfinished-adapter failures.
- [x] 2.2 Add a compile-tested public consumer registration example, package docs, license/source notices, and authoring guide explaining the host-adapter boundary; verify the generated package builds externally without Dome internal imports.

## 3. Maintainer assets and checks

- [x] 3.1 Generate baseline OpenSpec config/specs, indexed OKF concepts with actual source provenance, repository rules, and version-matched module-authoring skill; verify local links and metadata and that no generated claim fabricates host acceptance.
- [x] 3.2 Embed pinned OpenSpec installation/setup/wrapper scripts, isolated configuration, seven Codex workflows and upstream license; verify generated workflow versions, executable modes, setup repeatability, and preservation of the module-authoring skill.
- [x] 3.3 Supply repository-local knowledge validator/tooling with attribution and parity coverage against canonical validation fixtures; verify external generated artifact checks accept valid bundles and reject broken links/specs without Dome internal imports.
- [x] 3.4 Supply `Taskfile.yml` with direct build/check/test/race/tool-setup/artifact tasks and developer/update guides; verify every documented `task` command in the generated fixture and assert no Makefile or Make delegation is generated and confirm authored source is preserved by the documented manual update process.

- [x] 3.5 Make generated README, authoring docs, and module-authoring skill use `go doc -all <generated-module-path>` and `go doc github.com/jani123/dome/core/resource` as API references; verify those commands in the generated external fixture, inspect expected exported symbols, and document that selected dependency source determines the output.

## 4. Container tests and development environment

- [x] 4.1 Generate pinned Ginkgo/Gomega/Testcontainers dependencies, a customizable Debian Dockerfile, and repository-local helpers with bounded readiness/execution, failure logs, and cleanup; verify a real fixture starts and terminates without unrelated host mounts or leaked resources.
- [x] 4.2 Add the clearly test-only file-backed adapter, integration probe, and compiled consumer agent; verify tagged real-container suites cover boundary effects, plan without mutation, apply, contents, repeated no-op, pruning, and unrelated-file preservation while the production adapter still fails explicitly.
- [x] 4.3 Add direct `task test:integration` and `task test:e2e` commands and Docker/customization guides; verify both commands run successfully with Docker, fail clearly without it, and default unit/race tasks remain Docker-independent.
- [x] 4.4 Generate project-adapted Devbox/devcontainer/Compose/Docker-in-Docker/proxy assets and valid locks with no hardcoded Dome workspace or shared cache names; verify configuration resolution, executable modes, source provenance, and no tracked credentials or local state.
- [x] 4.5 Add included devcontainer lifecycle tasks, `task dev:setup`, post-create preparation, and environment/network customization docs; verify an actual generated environment builds, starts, passes toolchain checks, and runs both Testcontainers suites through its nested Docker daemon and proxy configuration.

- [x] 4.6 Generate minimal Hugo guide assets, pinned tooling, `task docs:build` and `task docs:serve`, and loopback-only random host port publication; verify the static build, served pages/assets, and simultaneous previews in two workspaces without port conflicts.
- [x] 4.7 Add project-scoped `task devcontainer:urls` and host/remote-Docker preview instructions; verify discovery lists only current published workspace services, labels the docs URL, distinguishes mappings from server readiness, handles stopped/recreated containers, and the reported docs address answers a real HTTP request.

## 5. Empty-directory publication

- [x] 5.1 Implement injected preparation and separate confined empty-directory publisher with staging, dependency resolution and no-replace publication; verify unit boundaries and integration-tagged real-filesystem tests for successful generation into the current directory.
- [x] 5.2 Add integration cases for hidden files, `.git`, symlinked ancestors, invalid paths, failed resolution, cancellation, concurrent creation and modified generated entries; verify unrelated content survives and private staging is cleaned on handled failures; verify partial published output is retained and reported rather than risking concurrent replacement deletion.
- [x] 5.3 Document multi-entry publication limits and interrupted-run recovery; verify an interrupted fixture is refused on rerun and cleanup guidance avoids deleting unrelated content.

## 6. CI and source releases

- [x] 6.1 Generate pinned GitHub CI installing the pinned Task version and invoking documented `task` commands for code/race/container integration/E2E/artifact/Hugo build and shell/workflow checks plus release scripts/workflow for supported immutable v0/v1 tags and draft source assets; verify actionlint, shell syntax, invalid-tag rejection, no agent-binary packaging, and no explicit pkg.go.dev indexing requests, availability polling, or hosted-documentation release gates.
- [x] 6.2 Add release notes/changelog starter and release guide covering job permissions, remote setup, public tag versus draft availability, version repair, and the distinction between public module retrieval and documentation indexing; verify all referenced assets exist and workflow gates/permissions match the guide.

## 7. Integrated verification and external acceptance

- [x] 7.1 Run the complete generated repository flow in isolated external storage, with an explicit local Dome replacement only for pre-release testing; verify build, unit/race/artifact checks, both real-container suites, and module registration example, and record that local replacement does not establish public resolution.
- [x] 7.2 Run `task check`, `task test:race`, `task test:e2e`, `task test:integration`, and `task check:artifacts` in Dome; verify both consumer and module initialization regressions and report any unperformed checks without marking them passed.
- [ ] 7.3 Exercise the published tagged initializer and generated checks natively on advertised Linux architectures without replacements; retain public module-resolution/platform acceptance as open until actual evidence exists.
- [ ] 7.4 Verify live Codex discovery and module/OpenSpec authoring in a generated repository; retain this acceptance as open until exercised rather than substituting front-matter checks.
- [ ] 7.5 In an explicitly authorized GitHub fixture, exercise successful and failing tag gates and inspect the draft source release and permissions; retain hosted acceptance as open until actual runs establish the behavior.

## Workflow follow-up

- Review implementation against scenarios, sync the new capability into baseline specs, and archive only verified completed scope. Keep unfinished external acceptance active.

## Acceptance ownership

This change owns generated module-repository assets and their tests. Core Taskfile migration, core API-guidance changes, and core docs preview are owned by their respective changes; this change consumes those conventions without duplicating core implementation. Module native bootstrap and live module-authoring skill checks remain distinct from existing consumer acceptance in verify-public-delivery. Reuse shared tagged Dome release/platform evidence where it establishes the same prerequisite, while separately exercising the new module output.
