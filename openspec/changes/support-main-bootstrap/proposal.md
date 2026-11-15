# Proposal

## Why

The development quickstart substitutes a hard-coded future release instead of the selected source version. Simply changing it to `@main` fails because initialization rejects the resolved Go pseudo-version.

## What Changes

- Render release bootstrap commands with the selected exact tag and development bootstrap commands with `@main`.
- Accept canonical v0 pseudo-versions resolved by Go for consumer initialization, pinning generated dependencies and matching metadata to that immutable version.
- Preserve rejection of unversioned local builds and mutable branch names as generated dependency pins.
- Keep resource-module initialization and release publishing policy outside this change.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- [project-initialization](../../specs/project-initialization/spec.md): permit immutable development pseudo-version pins for consumer projects.
- [documentation-publishing](../../specs/documentation-publishing/spec.md): map next guides to main and release guides to the selected tag.

## Impact

Affects initializer build-version detection, consumer scaffold validation/tests, the version shortcode, guide build regression checks, README, and curated versioning knowledge. No runtime reconciliation or persisted-state changes. Public main resolution and generated-consumer execution require actual acceptance checks.
