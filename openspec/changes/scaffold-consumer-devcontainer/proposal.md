# Proposal

## Why

The consumer initializer does not create a devcontainer, so the quickstart's temporary Docker shell cannot lead into the user's intended editor/CLI workflow. Generated projects need their own reproducible Go environment, and the entire walkthrough needs actual Dev Container CLI verification.

## What Changes

- Scaffold a consumer-owned `.devcontainer/devcontainer.json` and Dockerfile with pinned Go, a named non-root user, and a stable mounted workspace.
- Keep Go tooling in the container and project files on the host; do not apply resources or require the Docker socket inside the generated workspace.
- Show separate Go/Docker bootstrap alternatives and Dev Container CLI/VS Code opening alternatives, defaulting to Go and CLI respectively.
- Run all subsequent build, test, plan, apply, and repeat steps in the generated devcontainer and document shutdown/re-entry.
- Verify the complete walkthrough with the real Dev Container CLI, while distinguishing local implementation checks from unpublished main and GUI acceptance.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- [project-initialization](../../specs/project-initialization/spec.md): generated consumer devcontainer and complete container quickstart.

## Impact

Consumer templates, rendering/integration tests, docs shortcodes and tab tests, README and curated initialization knowledge. No resource-module devcontainer changes, runtime reconciliation changes, state migration, or application of generated resources during initialization.
