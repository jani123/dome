# Proposal

## Why

Core development should use Taskfile directly and consistently without a second task runner.

## What Changes

- Replace remaining Task-to-Make wrappers with direct commands or owned scripts.
- Migrate CI and maintained usage docs to `task` and remove the root Makefile after parity verification.
- Preserve versioned initializer build parameters and all required checks.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `development-environment`: Core development should use Taskfile directly and consistently without a second task runner.

## Impact

Taskfile.yml, Makefile, CI, README/developer docs, applicable repository rules and curated knowledge; compiled agent behavior unchanged.
