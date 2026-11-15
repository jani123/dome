# Proposal

## Why

The quickstart currently requires host Go and an ambiguously named module flag. A Docker shell and an explicit project-module flag should let users complete it without installing Go on the host or mistaking the project identity for an added resource dependency.

## What Changes

- **BREAKING** Rename the required consumer flag to `--project-module`; do not retain an alias for the old identity flag, and preserve module validation.
- Add repeatable `--module <module-path>@<exact-version>` inputs that import and register the root package as `Module{}`; users supply any constructor/backend parameters in generated Go.
- Keep resource-module initialization's required module argument unchanged.
- Add an optional pinned official Go Docker shell to quickstart and README, mounting the working directory and running as the caller's UID/GID. All subsequent quickstart commands run in that shell.
- Explain that the argument names the generated project and does not need to exist online.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- [project-initialization](../../specs/project-initialization/spec.md): explicit consumer project identity and Docker bootstrap guidance.

## Impact

Consumer initializer CLI, regression and real-container checks, quickstart, README, generated instructions, and curated knowledge. No release publishing, runtime, state schema, or resource-module CLI changes. Older tagged initializers retain their old flag name; their immutable docs remain unchanged.
