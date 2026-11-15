# Design

## Context

The consumer renderer embeds project assets but has no development container. The standalone resource-module scaffold has a much larger repository environment that is unnecessary for a simple consumer. Existing docs tabs support multiple independent groups, but tests currently assume one group.

## Goals / Non-Goals

Provide a small consumer-owned Go workspace and verify the requested end-to-end workflow. Do not copy maintainer tooling, add a module-scaffold runtime, install host Go, or mount the host Docker socket.

## Decisions

- Use official `golang:1.26.11-bookworm` as the Dockerfile base and create a named non-root `dev` user. Dev Container tooling aligns its UID/GID with the local user. Set both remote and default user consistently.
- Bind the project at `/workspace` using `${localWorkspaceFolder}` and document running subsequent commands there. Keep generated resource identity stable across starts; do not initialize or apply automatically.
- Forward standard proxy environment variables by reference, not embedded values. Restricted proxy hostnames may require an explicit test-local host mapping; document this configuration rather than substituting runtimes or bypassing blocked downloads.
- Docker bootstrap runs a one-shot initializer, returning to the host before devcontainer startup. Keep temporary container paths separate from the persistent workspace and avoid instructions to start Docker from inside a Go-only shell.
- Add a second independent tab group for CLI and VS Code. Keep the Go/bootstrap and CLI/open defaults independent, with accessible selection, panel switching, and keyboard support.
- Add an explicit acceptance command using the actual Dev Container CLI. Exercise Docker bootstrap from current source with a published library pin, generated container startup, the exact guide example, build/test, init/plan, first apply, no-op repeat, re-entry, ownership, and cleanup. Distinguish the local bootstrap variation from public-main acceptance.

## Risks / Trade-offs

- The new template is unavailable through public main until published → validate local sources without claiming the unpushed command has passed.
- Proxy-only environments cannot resolve the proxy hostname in a default Docker bridge → use explicit routing configuration and retain failure diagnostics.
- VS Code UI is not available in automated CLI verification → validate the shared configuration and documentation, and report GUI interaction as unexercised.

## Migration Plan

Existing projects remain untouched. Generate a temporary matching project and copy/review its development-container assets if adopting them. The new assets do not change the agent's state format or applied behavior.
