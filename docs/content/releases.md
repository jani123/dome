---
title: Versions and upgrades
weight: 70
---

One exact tag versions the Go library, initializer, embedded templates, skill,
examples, and guide source. Begin with `v0.1.0-rc.1`, then `v0.1.0` after the
acceptance checks. Prereleases are for explicit evaluation and never replace the
stable documentation alias. Releases are milestone-driven, with fixes as needed.

During v0, patch releases preserve public API, CLI/JSON contracts, and supported
state. Breaking changes require a minor release, migration notes, and explicit
consumer action. Only the latest minor receives fixes initially. v0 remains
experimental. Forward state migration does not promise binary downgrade support.

To update, commit your current source and back up state with the agent stopped.
Generate a temporary sibling project with the target exact initializer version.
Compare its dependency, skill, instructions, and `dome.json` with your project.
Update matching assets together and preserve custom resources/playbooks. Run
`go mod tidy`, build/test, inspect JSON plans, and then deploy deliberately.
No command automatically overwrites authored infrastructure code.

Published guide snapshots live under `/vX.Y.Z/`, development guides under
`/next/`, and the current stable copy under `/latest/`. Old snapshots remain
available. A published module tag is immutable; a failed release is repaired by
a new version, not moving the old tag. Unusable versions may be retracted.

The quickstart command follows the selected guide version: a release tag for
release guides, and `@main` for `next`. Go resolves main to an immutable
pseudo-version; the generated consumer dependency, skill, templates, and manifest
all use that same version. Development snapshots are experimental. Use local
`go doc` for the exact API of a pinned snapshot, since next guides follow main.
Resource-module initialization continues to require an exact release tag.
