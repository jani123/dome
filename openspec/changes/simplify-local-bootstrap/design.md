# Design

## Context

Consumer CLI passes an empty required module to scaffold validation today. The scaffold already supports local Go module identities with a dotted first component; Go does not fetch the main consumer module. The generated project has no devcontainer. See proposal.md for motivation.

## Goals / Non-Goals

Provide container-only quickstart tooling and an explicit, understandable consumer identity. Do not generate a devcontainer, change module init, or change reconciliation, state, and release behavior.

## Decisions

- Require `--project-module` for consumer init. Do not treat `--module` as an alias for the project identity. Validate before any publication; do not infer or silently normalize names.
- Keep the existing Render entry point unchanged; add a renderer accepting external module selections. Separate module-command flag storage from consumer flag storage so reusing a Cobra command cannot leak a consumer identity into resource-module initialization.
- Use an optional interactive shell in official `golang:1.26.11-bookworm`, mounted at `/workspace`, with UID/GID and writable temporary HOME. Run the entire walkthrough in that shell so resource identity stays consistent. No Docker socket, privileged mode, host Go, or downloaded host binaries are needed. Tool caches are ephemeral.
- Preserve selected-version command substitution. Immutable old release docs and initializers retain the old flag requirement; current next docs describe the renamed command. Local container acceptance may use a checkout replacement, explicitly distinguished from public main resolution.

- External selections use root Go module paths and canonical exact versions. Validate paths/versions and reject duplicate imports, self-imports, and Dome's built-in module path before publication. Generate deterministic aliases and `Module{}` registrations in Go, plus direct go.mod requirements. Root packages must export a Module type; users edit registrations needing parameters. Do not claim these zero-value registrations are configured or usable without that work.
- Run normal dependency resolution in the staging directory, then compare the requested Dome and external module pins against resolved go.mod. Refuse publication if Go selects different versions or resolution fails; preserve causes without exposing command output. Do not compile or apply while initializing, since user-supplied registration parameters may still be required.

## Risks / Trade-offs

- Older releases do not understand the new flag → release guide builds retain `--module`; next uses `--project-module`.
- Bind mounts require a host-local Docker daemon with permission to the working directory → document Linux Docker access and exercise a real isolated bind mount.
- Network/proxy issues may prevent image/module retrieval → report actual endpoints and errors; do not substitute dependencies.

## Migration Plan

Existing projects remain valid. Consumer creation now requires the renamed identity flag. Roll back CLI help and next guide commands together if needed. No persisted-state migration.
