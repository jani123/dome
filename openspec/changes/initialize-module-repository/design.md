# Design

## Context

See [proposal](proposal.md) for motivation. `cmd/dome/main.go` currently exposes only consumer `init` and `version`. `internal/scaffold/scaffold.go` renders embedded consumer templates and a matching version manifest; `publish.go` resolves dependencies in a sibling staging directory before Linux no-replace rename. It rejects existing destinations and always adds a sandbox directory, so it cannot directly publish a module into the current directory.

Public `core/resource` interfaces already support external modules, including `Transition` for explicit safe defaults. `internal/knowledge` provides the repository validator but Go internal visibility prevents generated external modules from importing it. Current OpenSpec setup pins 2.0.0 and generates seven Codex workflows. These maintainer assets are currently absent from consumer starters.

## Goals / Non-Goals

**Goals:** produce an independently usable resource library repository, with reproducible developer assets and honest unfinished host behavior. Keep module generation separate from consumer generation while reusing pure validation/rendering primitives.

**Non-Goals:** change reconciliation APIs, implement a particular host integration, merge into populated repositories, upgrade generated projects automatically, configure GitHub settings, or add a documentation hosting platform.

## Decisions

### CLI and version contract

Use `dome module init --module github.com/acme/dome-widget --name widget` in the current directory. Both flags are required; the name is a lowercase non-keyword Go identifier shared by the root package and resource module identity. Validate the external module path with Go module tooling and exclude Dome's own module namespace. Use the initializer's supported exact v0 release as the dependency pin and module skill/template version. Reject dev metadata as consumer init does. Initial release guidance supports v0/v1 tags; higher major module-path migration remains manual and unsupported by the starter release gate.

No directory positional argument or modes in this first version: the user's empty-working-directory model is sufficient. Preserve `dome init` semantics unchanged.

### Generated repository inventory

Generate a root library package (`module.go`, `resource.go`, `backend.go`), `doc.go`, tests using Testify, and a compile-tested explicit `agent.WithModules` example. Export `Module` and `Resource` and a narrow `Backend` interface with Observe and Apply methods; supply an explicit error-returning unimplemented adapter. Use stable non-secret names, contextual fault categories, validated `Restore`, action checking, and `Transition` safe-default declarations. Document that real adapters must provide truthful observation, safe repetition/deletion, context handling, and non-secret snapshots before use. Tests cover the adapter contract with fakes; they do not certify any host integration.

Include `go.mod`/`go.sum`, exact template manifest, `.gitignore`, GPL-3.0-only license and required source notices, `AGENTS.md`, README, and plain Markdown `docs/authoring.md`, `docs/development.md`, and `docs/releases.md`. Licensing follows the existing starter rather than introducing a new license choice. Explain explicit consumer registration and separate module release versions from the pinned Dome version.

Include `openspec/config.yaml`, no fabricated active or archived changes, and initial specs describing only the generated contract and unfinished adapter. Include `knowledge/index.md` and indexed architecture/decisions/operations concepts referencing real generated sources, without copied Dome historical verification metadata. Specs own requirements, knowledge owns rationale/provenance, docs own commands. Add a version-matched module-authoring Codex skill alongside the seven OpenSpec skills.

### Local API documentation

Use Go source comments, compile-tested examples, and local `go doc` as the API reference. Generated README and authoring skill instructions show `go doc -all <generated-module-path>` and `go doc github.com/jani123/dome/core/resource`, run from the module or consumer workspace so documentation reflects the selected source/dependency version. Document that dependency retrieval may require network access before local inspection; no hosted documentation index is needed.

Generated workflows do not explicitly request pkg.go.dev indexing, poll its pages, or gate release acceptance on hosted documentation availability. Continue verifying public tagged module resolution: that is a distribution check, not a documentation indexing check. Public proxy retrieval may itself cause indexing, so neither docs nor tooling claim that indexing is prevented. Hiding current/future documentation requires a separate request to pkg.go.dev maintainers, as described in their [removal policy](https://pkg.go.dev/about#removing-a-package); initialization does not submit such requests.

### Self-contained tooling and asset provenance

Embed a separate module template tree. Vendor a deliberately small copy of the existing knowledge validator and its tests into generated repository-local tooling, retaining license/source attribution. Do not export a new public framework package merely to bootstrap a developer validator, and never import Dome `internal` paths. Maintain the copied template through parity tests against the canonical validator to prevent drift.

Provide `Taskfile.yml` with `build`, `check`, `test`, `test:race`, `tools:openspec`, `specs:setup`, `check:knowledge`, `check:specs`, and `check:artifacts` tasks. Implement commands directly in tasks or repository-owned scripts; generate no Makefile and no delegation to Make. Use `task` consistently in generated README/guides, authoring skills, and GitHub workflows, installing a pinned Task version in CI and documenting local installation prerequisites. Include the pinned native OpenSpec installer, isolated wrapper/configuration, workflow version checker, seven generated workflows and upstream MIT notice. Generate/refresh those assets in staging using the pinned CLI; keep tool caches/config in ignored storage and include required executable modes in the rendered file model. Document Linux amd64/arm64 and Go prerequisites consistently with current initializer support. The module is not an agent and needs no runtime sandbox or state database.

### Empty-directory publication

Keep the consumer publisher separate. Open the destination as a confined directory handle, validate real ancestors and initial emptiness including dotfiles, and record its identity. Render, install/setup developer assets, and resolve Go dependencies in owned sibling staging before publishing. Never move or replace the destination directory: it can be the process's current working directory.

Recheck destination identity and emptiness, then publish top-level entries with no-replace operations relative to the retained destination handle. Track published entry identities. On a handled error or cancellation, remove only owned entries proven unchanged; leave changed or concurrently added content untouched and report incomplete cleanup. Do not recursively delete content whose ownership cannot be established. Validate template paths before staging. Errors exclude subprocess output that might reveal private module/proxy credentials.

Multi-entry publication into an existing directory is not fully atomic. On process death, output can remain partial; docs explain inspection and manual cleanup or starting in a new empty directory, and reruns refuse partial output. Promise no overwrites, not crash rollback. Replacing the directory atomically was rejected because it breaks the current-directory contract; directly writing templates was rejected because preparation failures would expose avoidable partial output.

### CI and release workflow

Generate CI on pull requests, main pushes, and reusable workflow calls: formatting/vet/unit, race, artifact validation, shell syntax and actionlint checks. Pin actions and developer tools consistently. Tag workflows accept canonical supported module versions, verify tag/module-path compatibility, and call the same checks before a dedicated minimal-write job creates a draft source release with changelog notes, source archive, checksums, and license notices. No cross-compiled agent binaries or successful-host-integration claims.

Use GitHub's job token, no user token requirement. Initialization only writes workflow files. Authors create the remote, enable Actions, push commits/tags, inspect drafts, and publish explicitly. Go proxies may resolve a pushed tag before draft publication; docs explain immutable tags and repair by new versions. Generated workflow lint and local tests are separate from live hosted acceptance.

## Risks / Trade-offs

- Partial output after abrupt termination → document recovery and refuse merges; test handled rollback and race cases.
- Copied tooling drift and larger embedded assets → retain provenance, parity tests, pinned versions, and regenerated-asset review.
- A compilable scaffold can appear production-ready → fail explicitly without an adapter, document the boundary, and avoid fabricated external acceptance.
- Release workflows require hosted permissions and public tags → keep live acceptance open until exercised in an explicitly authorized fixture repository; local generation never performs remote writes.
- Template additions could disturb consumer init → retain its golden/version tests and run both independent generated-repository flows.

## Migration Plan

Add the module subcommand and independent templates without rewriting existing generated projects. Update Dome README/authoring docs and curated knowledge with the new starter's boundaries, and sync the capability only through the completed OpenSpec lifecycle. Run generated module checks in a temporary external module and preserve outstanding public-tag/Codex/GitHub acceptance as active tasks. Before release, rollback is removal of the new subcommand/templates; generated repositories remain author-owned and updates use reviewed temporary regeneration.
