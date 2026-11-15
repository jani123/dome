## MODIFIED Requirements

### Requirement: Pinned guide build

Guides SHALL build with pinned checksum-verified Hugo and checked internal paths under /dome/; API reference guidance SHALL use local `go doc` against the selected checkout or dependency version, without requiring pkg.go.dev indexing or availability. Release bootstrap commands SHALL use the selected guide tag; next bootstrap commands SHALL use main.

Evidence: [scripts/build-docs.sh](../../../scripts/build-docs.sh).

#### Scenario: Pinned guide build

- **WHEN** an RC guide is built
- **THEN** its bootstrap command uses that RC version and API guidance explains selecting that version locally before running `go doc`

#### Scenario: Hosted API index unavailable

- **WHEN** pkg.go.dev is unavailable or has not indexed the selected version
- **THEN** users can inspect that version with local `go doc` after its source is available and no release gate waits for hosted documentation

#### Scenario: Development bootstrap

- **WHEN** next guides are rendered
- **THEN** the quickstart initializer command uses `@main` without substituting a future release tag
