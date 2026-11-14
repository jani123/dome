# Proposal

These artifacts were recorded after implementation. They document the delivered scope and do not represent prior planning or approval.

## Why

Core contributors need the same local docs preview and URL discovery planned for module repositories.

## What Changes

- Serve pinned Hugo with Task and publish a random host-loopback port.
- Discover current workspace mappings without starting containers.
- Document preview access and validate real HTTP behavior.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `development-environment`: Discoverable devcontainer docs preview.

## Impact

Taskfiles, devcontainer Compose, preview/discovery scripts, maintainer docs and knowledge. Agent runtime remains unchanged.
