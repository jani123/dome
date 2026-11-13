# Module initialization

## Purpose

Bootstrap an independently maintained Dome resource-module repository with working Go contracts, specifications, curated knowledge, documentation, and release tooling. These requirements describe proposed behavior, not current implementation or completed external acceptance.

## ADDED Requirements

### Requirement: Explicit standalone module inputs

`dome module init --module <go-module-path> --name <resource-module-name>` SHALL initialize the current existing empty directory. It SHALL validate an external Go module path, a non-keyword lowercase Go package identifier as the resource-module name, and an exact supported Dome release version before publishing files.

#### Scenario: Initialize a module
- **WHEN** the command runs in an empty real directory with valid inputs using a supported tagged initializer
- **THEN** it creates a standalone Go library with the requested import path and module identity and reports development next steps

#### Scenario: Invalid metadata
- **WHEN** required inputs are absent, names or paths are invalid, or the initializer has unsupported or development version metadata
- **THEN** it fails with the CLI validation exit code and leaves the destination empty

### Requirement: Safe publication in an empty directory

Initialization SHALL refuse any preexisting entry, including hidden files and `.git`, and symlinked destination paths or ancestors. It SHALL stage and resolve dependencies before publication, never overwrite concurrent entries, and clean only its own unchanged output after handled failure or cancellation. It SHALL document interrupted-publication recovery.

#### Scenario: Directory already contains files
- **WHEN** the directory contains even one hidden or visible entry
- **THEN** initialization fails without modifying its contents

#### Scenario: Resolution fails
- **WHEN** dependency resolution or staged preparation fails
- **THEN** the destination remains empty and owned staging output is removed

#### Scenario: Another writer creates a collision
- **WHEN** another process creates an entry before that generated entry is published
- **THEN** initialization fails without replacing or deleting the other process's entry

### Requirement: Public and safe resource starter

The generated library SHALL implement the public resource, module, and safe-default transition contracts with injectable external operations and mock-based tests. It SHALL validate identities, restore deletion handles without old secrets, and use non-secret snapshots. Unimplemented host behavior SHALL fail explicitly rather than report successful observation or mutation.

#### Scenario: Exercise the generated package
- **WHEN** an author runs the generated checks
- **THEN** the library and registration example compile and tests exercise validation, restoration, observation, apply actions, and boundary failures through injected fakes

#### Scenario: Use the unfinished adapter
- **WHEN** a caller observes or applies a resource without a real host adapter
- **THEN** it receives a categorized error rather than a fabricated converged state

### Requirement: Ready repository authoring assets

The generated repository SHALL include OpenSpec configuration, baseline specs for the generated scaffold, pinned OpenSpec tooling and seven Codex workflows, an OKF v0.2 indexed bundle with source provenance, repository instructions, README, authoring and release guides, and a module-authoring skill matched to its Dome dependency. No knowledge claim SHALL assert unperformed verification.

#### Scenario: Develop a new resource behavior
- **WHEN** an author follows generated instructions
- **THEN** they can locate source-linked knowledge, propose and implement an OpenSpec change, and run spec and knowledge validation without importing Dome internal packages

### Requirement: Versioned self-contained developer tooling

Generated build, unit, race, and artifact checks SHALL work in an external repository without a Dome source checkout or local replacements. Dependency, module-authoring skill, and template metadata SHALL match the initializer's exact Dome release. Tooling SHALL remain outside deployed agent runtime, with documented manual updates preserving authored code.

#### Scenario: Validate a clean external checkout
- **WHEN** the documented tools are installed and checks run in a generated repository
- **THEN** its Go code, OpenSpec artifacts, and OKF bundle validate using pinned dependencies and tools

### Requirement: Taskfile development workflow

Generated repositories SHALL use `Taskfile.yml` as their sole task runner, with direct task or repository-script implementations and no Makefile or Make dependency. Generated documentation, authoring skills, and GitHub workflows SHALL use consistent `task` commands and document or install the required Task version.

#### Scenario: Run documented development commands
- **WHEN** an author runs the documented build, test, race, or artifact tasks with the documented tools installed
- **THEN** the tasks execute without requiring Make

#### Scenario: Run generated CI
- **WHEN** a generated GitHub check job starts
- **THEN** it installs the pinned Task version and invokes the same task commands documented for local development

### Requirement: Local API reference

Generated documentation and authoring skills SHALL direct users to local `go doc` for the selected module and Dome dependency, with package comments and examples as API documentation sources. Workflows SHALL NOT explicitly request pkg.go.dev indexing or depend on its availability. Documentation SHALL NOT claim public module indexing is prevented.

#### Scenario: Inspect the selected API
- **WHEN** an author follows generated API-reference commands in the prepared module or consumer workspace
- **THEN** `go doc` displays the selected local or dependency API without consulting pkg.go.dev

#### Scenario: Release independently of hosted documentation
- **WHEN** pkg.go.dev has not indexed a version or is unavailable
- **THEN** API inspection and release checks remain usable without requesting or waiting for indexing

### Requirement: Check-gated library releases

Generated GitHub workflows SHALL run code, race, artifact, and workflow checks on pull requests and main-branch pushes. Supported immutable version tags SHALL run the same gates before preparing a draft source-library release. Actions SHALL be pinned to immutable commits and write permission limited to the draft job. Documentation SHALL distinguish tagged Go module availability from GitHub Release publication.

#### Scenario: A release check fails
- **WHEN** a tag-triggered required check fails
- **THEN** the workflow does not create a draft release

#### Scenario: Prepare a module release
- **WHEN** a supported version tag passes all required checks
- **THEN** the workflow creates a draft with release notes and source licensing material and no agent binaries

### Requirement: Local generation only

Initialization SHALL NOT initialize Git, create remote repositories, push commits or tags, publish releases, register a module into an agent, or execute host resource changes. Existing `dome init` behavior SHALL remain compatible.

#### Scenario: Complete initialization
- **WHEN** generation succeeds
- **THEN** only local repository files and required staged tool/dependency preparation have occurred, and remote setup and host integration remain explicit author actions
