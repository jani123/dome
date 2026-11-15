# Design

## Context

See proposal.md. cli.New currently builds commands around an injectable Backend; Options carries source and state overrides but no presence information for defaulted flags. Agent.New defaults to an empty local source and .dome/state.db; agent.evaluate selects explicit Git/local options over a.source. The generated consumer embeds a local config.yaml source and a relative sandbox resource identity. Configuration types belong to consumer Go code and cannot be reconstructed from module registration.

## Goals / Non-Goals

**Goals:** Repeatable operator bootstrap, deterministic saved defaults, prompt-free service operation, compatibility with existing Go composition, and protected state identity.

**Non-Goals:** Go project generation, remote Git commits/clones, key generation or host-key enrollment, secret entry/storage, filesystem-wide profile discovery, introspection of arbitrary desired-state functions, or automatic resource reconciliation during setup.

## Decisions

### Workspace and profile boundary

The current directory is the default workspace, as selected by the owner. Add persistent --workspace <directory>. Do not search parents or home directories. Save .dome/agent.yaml with schema_version: 1 and explicit source kind (local/git), source fields, secret-file reference and state path. --agent-config <path> selects a complete bootstrap profile for init or a runtime invocation, independently from --config (resource YAML). The external profile's relative values resolve against --workspace/current directory; paths are normalized into the saved profile without copying credential files. Explicit missing profile paths fail; absent default profiles preserve legacy Go/flag workflows.

Keep profiles outside the SQLite schema. Init creates directories and runtime profile output only; it never opens state, evaluates desired resources or calls Apply. A profile is not proof of source connectivity, successful SOPS decryption or valid resource values. Normal source loading remains authoritative. Profiles are permission-restricted (directory 0700, files 0600).

### Input and prompting

Collect candidate values from explicit flags over the supplied profile; track presence separately from Cobra defaults. Prompt on terminal input only for missing required values, using cmd input and diagnostic output streams through injectable prompt/terminal interfaces. Select source kind before source-dependent fields. Defaults offered explicitly include Git ref HEAD, blob config.yaml and state .dome/state.db; HTTP requires deliberate opt-in. SSH requires existing readable key/known-hosts file references. Local setup selects an existing configuration file. Ask for optional encrypted-secret reference with an explicit none choice. Never request plaintext secrets or key contents. --non-interactive, --json and non-terminal input never prompt; complete input succeeds and missing input returns validation errors. Cancellation/EOF terminates setup cleanly. No prompts run in plan/apply/run/status or diagnostic commands.

### Host runtime settings, not repository scaffolding

The saved .dome/agent.yaml is the runtime configuration scaffold. For Git-backed
setup, create no local resource YAML, checkout, example file or commit instructions.
Resource configuration remains authored in the existing Git repository. Local-source
setup requires an existing resource configuration file and preserves it. Do not
add a schema-inference or consumer-template API for this feature. The existing
source-generation commands remain dome init and dome module init; a dedicated
configuration-only Git repository initializer is not part of this change.

### Help and command flag placement

Advertise --config, --secrets, --git-url, --git-ref, --git-path,
--git-allow-http, --ssh-key, --known-hosts and --state only in init help.
Keep --workspace, --agent-config and applicable output/command-specific flags
visible for runtime commands. Existing setup flags are currently persistent root
flags: replace that registration with visible init-local flags and hidden legacy
registrations on commands that already accept them, preserving accepted flag
positions and override semantics. Do not toggle a shared flag's hidden state
based on which help command ran: fresh or reused command trees must behave
consistently. Root/global/inherited help must not expose hidden setup flags.
Normal README, generated consumer and configuration/deployment examples show
init once then plan/apply/run with saved settings. Document retained overrides
only in compatibility guidance. Verify help output on root and nested runtime
commands as well as init, including inherited/global sections.

### Command resolution and compatibility

For runtime source/state commands, resolve flags > selected profile > Go defaults. Read profile only for commands requiring runtime settings; help/version/modules/playbook list remain available with a damaged profile. Resolve once per invocation, including a polling run. Use flag Changed to distinguish omitted defaults from explicit false/empty values. Explicit --config or --git-url selects a coherent source; never carry secret paths from the old source into a new one. Source-specific ref/path flags may refine the saved Git source; reject Git-only fields with local sources. Explicit --secrets belongs to the effective selected source. Explicit --state overrides saved state, which otherwise overrides WithSQLite. Never rewrite a profile during runtime commands.

Anchor profile paths to the workspace using absolute resolved values passed through Options. Do not os.Chdir or reinterpret relative resource identities in arbitrary Go callbacks. Existing generated sandbox topology remains tied to its documented working directory; docs must distinguish runtime-settings selection from changing consumer-defined resource paths. Scaffold/example updates must keep resource identity stable in their supported invocation flow. Selecting a workspace does not migrate existing state or recover a discarded database.

### Safe publication and repeat setup

Validate all settings before writing. Use injectable filesystem operations, anonymous O_TMPFILE staging and atomic no-replace publication of the pinned written inode through /proc/self/fd, rejecting symlinked destinations/ancestors and concurrent replacements. Check the selected directory identity before and after linking. Never unlink a visible published path on subsequent failure. Anonymous staging avoids mutable source names and pathname cleanup. The Linux filesystem must support O_TMPFILE and hard links, and /proc/self/fd must be available; unsupported capabilities return concealed errors preserving their causes. Enforce the same 64 KiB profile bound before publication and during loading. Create only missing workspace directories and never change ownership or permissions of unrelated existing paths. Failures leave no initialized profile; cleanup is limited to owned staging files. Equivalent existing profile/settings is a no-op; conflicting init fails with guidance to edit a profile deliberately, without force/overwrite functionality in this scope. Existing state is never reset.

### Verification and maintained documentation

Mock filesystem, prompts, terminal detection and backend calls in unit tests. Add integration-tagged real filesystem, permission, symlink and concurrency cases. Extend compiled-agent container tests with init then repeated apply, state diagnostics, idempotency and failed initialization, plus real Git SSH/HTTP retrieval from saved settings. Keep all recovery/revision invariants in the existing suites. Update README, configuration/deployment guides, generated consumer source/docs and relevant curated knowledge with actual verified behavior. No release/public-delivery acceptance claim follows from local tests.

## Risks / Trade-offs

- Hidden defaults could switch source or state → explicit presence tracking, source-unit replacement and tests capturing Backend Options.
- Relative Go resource paths remain consumer-owned → no global chdir and explicit identity documentation/testing.
- Interrupted publication → staged atomic no-replace profile publication and cleanup of owned staging files.
- Interactive setup might hang services or corrupt JSON → terminal detection, non-interactive/JSON fail-fast and diagnostics on stderr.
- New init must not hide existing recovery history → profile initialization never opens/resets state; default state remains .dome/state.db.

## Migration Plan

Existing explicit invocations continue to work. Operators can pass their existing flags once to init, select the existing state path, inspect the non-secret profile and then use plan/apply without repeated flags. Keep the prior state database and resource identities. Removing a profile returns to legacy Go/explicit settings; it does not undo resource changes. No SQLite migration is needed.
