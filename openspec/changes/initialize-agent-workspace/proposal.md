# Proposal

## Why

Compiled agents currently require repeating Git source and credential-path flags for each invocation. Operators need a one-time workspace setup that saves non-secret runtime settings, creates local runtime configuration and asks for missing inputs interactively.

## What Changes

- Add compiled `agent init`, distinct from source-project `dome init` and library `dome module init`.
- Persist a versioned, strict YAML workspace profile for local/Git configuration sources, secret-file references and state location.
- Initialize the current directory by default with `--workspace` selecting another directory. Create only local runtime settings; Git-backed resource configuration remains in its existing repository.
- Accept a bootstrap profile via `--agent-config` and existing source flags; prompt only for missing values in interactive init. Support complete unattended setup with `--non-interactive`.
- Advertise source, credential, secrets and state setup flags on init only. Runtime help and examples use saved workspace settings; accept legacy override flags without showing them in runtime/root help.
- Reuse saved settings for subsequent runtime commands; explicit flags override saved defaults for one invocation. Preserve existing uninitialized explicit-flag and Go-configured workflows.
- Preserve state identity, verified SSH, explicit HTTP opt-in and secret handling. Initialization does not reconcile, clone/push a remote repository, create keys or trust unknown host keys.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `consumer-cli`: Workspace initialization, prompting and saved invocation defaults; see [current spec](../../specs/consumer-cli/spec.md).
- `configuration-sources`: Strict bootstrap profile and deterministic source/path precedence; see [current spec](../../specs/configuration-sources/spec.md).

The [project initialization](../../specs/project-initialization/spec.md) source-generation contract remains separate; generated consumers receive compatible runtime setup examples.

## Impact

cli command construction and injectable I/O/filesystem boundaries, agent command composition, consumer scaffold source/docs, example agent, configuration profile loading, CLI/unit/integration and compiled-agent E2E tests. No change to resource topology, reconciliation ordering, migrations or applied revision semantics. The owner selected the current directory with a workspace override and settings-only host initialization for Git sources.
