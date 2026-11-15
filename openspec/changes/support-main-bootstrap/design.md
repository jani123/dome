# Design

## Context

See proposal.md for motivation. The shortcode currently substitutes v0.1.0 for next. Build metadata detection and consumer scaffold validation both reject pseudo-versions. Generated dependency, skill, template, and manifest versions must agree.

## Goals / Non-Goals

Support consumer bootstrap from main without weakening immutable version matching. Resource-module initialization, release-tag policy, and runtime behavior remain outside scope.

## Decisions

- Use Go's resolved build metadata as the development version source. Accept canonical v0 pseudo-versions for consumer generation while preserving current version checks otherwise. Do not persist `main`: it would allow dependency resolution to drift from embedded assets.
- Render `main` for next and the existing selected version for tagged guide builds. Do not invent a development release tag or use latest, which would select a different source.
- Keep resource-module scaffold validation release-only. Shared build metadata detection can expose the resolved version, but that command retains its own version validation.
- Extend rendered-guide checks for next, stable, and RC builds. Add initializer and scaffold regression coverage for canonical pseudo-versions, invalid inputs, metadata agreement, and unchanged exact-tag behavior.

## Risks / Trade-offs

- Main is mutable and may be unstable → document development status and pin each generated project to Go's resolved immutable version.
- Public main has not yet received the implementation → distinguish local verification from public acceptance; public acceptance must wait until the change is reachable through main.
- Some version checks elsewhere may assume release-only inputs → follow metadata consumers and update consumer guidance consistently without broadening release publishing.

## Migration Plan

Existing consumers and tagged snapshots need no migration. Update next after deployment; existing generated consumers retain their pins. Roll back next guidance and initializer changes together if needed. No state migrations.
