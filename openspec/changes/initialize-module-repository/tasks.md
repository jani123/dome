# Tasks

## 1. Command and render contract

- [ ] 1.1 Add module-init command inputs, exact version validation, and pure module rendering; verify mocked command tests cover required flags, invalid paths/names/versions, help, exit categories, and unchanged consumer init.
- [ ] 1.2 Add the module template inventory, provenance/license metadata, and executable-file modes; verify render tests check formatted Go, exact version agreement, safe relative paths, expected files, and no consumer sandbox assets.
- [ ] 1.3 Document invocation, empty-directory prerequisite including `.git`, and supported platforms in Dome README/authoring docs; verify examples match CLI help and update source-linked curated knowledge.

## 2. Resource library starter

- [ ] 2.1 Implement templates for public Module, Resource, Backend, categorized unfinished adapter, validated restore and safe-default transition; verify generated Testify tests cover identity, restoration, snapshots, all actions, cancellation/error propagation, and unfinished-adapter failures.
- [ ] 2.2 Add a compile-tested public consumer registration example, package docs, license/source notices, and authoring guide explaining the host-adapter boundary; verify the generated package builds externally without Dome internal imports.

## 3. Maintainer assets and checks

- [ ] 3.1 Generate baseline OpenSpec config/specs, indexed OKF concepts with actual source provenance, repository rules, and version-matched module-authoring skill; verify local links and metadata and that no generated claim fabricates host acceptance.
- [ ] 3.2 Embed pinned OpenSpec installation/setup/wrapper scripts, isolated configuration, seven Codex workflows and upstream license; verify generated workflow versions, executable modes, setup repeatability, and preservation of the module-authoring skill.
- [ ] 3.3 Supply repository-local knowledge validator/tooling with attribution and parity coverage against canonical validation fixtures; verify external generated artifact checks accept valid bundles and reject broken links/specs without Dome internal imports.
- [ ] 3.4 Supply `Taskfile.yml` with direct build/check/test/race/tool-setup/artifact tasks and developer/update guides; verify every documented `task` command in the generated fixture and assert no Makefile or Make delegation is generated and confirm authored source is preserved by the documented manual update process.

- [ ] 3.5 Make generated README, authoring docs, and module-authoring skill use `go doc -all <generated-module-path>` and `go doc github.com/jani123/dome/core/resource` as API references; verify those commands in the generated external fixture, inspect expected exported symbols, and document that selected dependency source determines the output.

## 4. Empty-directory publication

- [ ] 4.1 Implement injected preparation and separate confined empty-directory publisher with staging, dependency resolution and no-replace publication; verify unit boundaries and integration-tagged real-filesystem tests for successful generation into the current directory.
- [ ] 4.2 Add integration cases for hidden files, `.git`, symlinked ancestors, invalid paths, failed resolution, cancellation, concurrent creation and modified generated entries; verify unrelated content survives and unchanged owned staging/output is cleaned on handled failures.
- [ ] 4.3 Document multi-entry publication limits and interrupted-run recovery; verify an interrupted fixture is refused on rerun and cleanup guidance avoids deleting unrelated content.

## 5. CI and source releases

- [ ] 5.1 Generate pinned GitHub CI installing the pinned Task version and invoking documented `task` commands for code/race/artifact and shell/workflow checks plus release scripts/workflow for supported immutable v0/v1 tags and draft source assets; verify actionlint, shell syntax, invalid-tag rejection, no agent-binary packaging, and no explicit pkg.go.dev indexing requests, availability polling, or hosted-documentation release gates.
- [ ] 5.2 Add release notes/changelog starter and release guide covering job permissions, remote setup, public tag versus draft availability, version repair, and the distinction between public module retrieval and documentation indexing; verify all referenced assets exist and workflow gates/permissions match the guide.

## 6. Integrated verification and external acceptance

- [ ] 6.1 Run the complete generated repository flow in isolated external storage, with an explicit local Dome replacement only for pre-release testing; verify build, unit/race/artifact checks and module registration example, and record that local replacement does not establish public resolution.
- [ ] 6.2 Run `task check`, `task test:race`, `task test:e2e`, `task test:integration`, and `task check:artifacts` in Dome; verify both consumer and module initialization regressions and report any unperformed checks without marking them passed.
- [ ] 6.3 Exercise the published tagged initializer and generated checks natively on advertised Linux architectures without replacements; retain public module-resolution/platform acceptance as open until actual evidence exists.
- [ ] 6.4 Verify live Codex discovery and module/OpenSpec authoring in a generated repository; retain this acceptance as open until exercised rather than substituting front-matter checks.
- [ ] 6.5 In an explicitly authorized GitHub fixture, exercise successful and failing tag gates and inspect the draft source release and permissions; retain hosted acceptance as open until actual runs establish the behavior.

## Workflow follow-up

- Review implementation against scenarios, sync the new capability into baseline specs, and archive only verified completed scope. Keep unfinished external acceptance active.
