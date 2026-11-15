## MODIFIED Requirements

### Requirement: Version-matched generation

Consumer initialization SHALL require a valid external consumer module and a canonical v0 release version or canonical v0 Go pseudo-version resolved for the initializer. It SHALL pin the dependency, skill, templates, and manifest to the same immutable version, rejecting unversioned dev builds, mutable branch names, and unsupported version inputs.

#### Scenario: Version-matched generation

- **WHEN** a development initializer lacks explicit candidate metadata
- **THEN** init refuses to create an unpublished dependency pin

#### Scenario: Main bootstrap

- **WHEN** Go resolves an initializer invoked with `@main` to a canonical v0 pseudo-version
- **THEN** consumer initialization succeeds with that exact pseudo-version in its dependency and matching generated version metadata

#### Scenario: Release bootstrap

- **WHEN** the initializer is invoked at an exact supported release tag
- **THEN** the consumer dependency and generated metadata retain that exact release version
