# Design

## Context

Taskfile already runs Go checks directly but delegates initializer/tool/spec checks to Make; CI still uses Make target names. Preserve initializer VERSION defaults and linker metadata, required suite flags/timeouts, formatting failure behavior, and artifact validation ordering when moving commands. Update workflow matrix entries to explicit Task names, install a pinned Task version, and update maintained AGENTS.md verification commands to agree. Archived historical commands remain evidence, not current instructions. Remove Makefile after task parity is verified. Coordinate with expose-core-docs-preview and module planning without claiming those unfinished changes are implemented.

## Risks / Trade-offs

These are proposed changes, not completed implementation or acceptance. Preserve unrelated work and historical evidence; update current guidance without rewriting archived history.
