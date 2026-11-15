---
title: YAML, Git, and secrets
weight: 40
---

Define topology in Go. Load values with `config.Parse[Settings, Secrets]` into
user-defined structs. Unknown fields, malformed configuration, retrieval, or
SOPS decryption errors must stop reconciliation and pruning.

Store secrets as SOPS-encrypted YAML. Provision decryption keys on the target,
for example with `SOPS_AGE_KEY_FILE`. Do not commit plaintext keys or secrets.
Keep decrypted values out of logs, error strings, resource names, snapshots,
and execution metadata. Customer source delivery must not publish secrets.

The starter reads `config.yaml` locally. Set up the managed host once:

```sh
./agent init
```

Initialization prompts for missing source settings on a terminal. It creates
`.dome/agent.yaml` in the current workspace, without applying resources, creating
SQLite state, cloning a repository or generating resource YAML. For unattended
Git setup, supply all required inputs:

```sh
./agent init --non-interactive --git-url ssh://git@git.example.com/configuration \
  --git-ref main --git-path config.yaml \
  --ssh-key /etc/my-agent/git.key --known-hosts /etc/my-agent/known_hosts
./agent plan --json
./agent apply --json
```

Verify SSH host keys through a trusted channel. Public HTTPS retrieval is
supported. Plain HTTP requires explicit `--git-allow-http`; never enable it
implicitly. Private HTTPS authentication, encrypted SSH private keys, submodules,
and LFS configuration blobs are not supported. Keep configuration repositories
small because retrieval fetches history afresh in memory.

With Git, `--secrets secrets.enc.yaml` reads encrypted secrets from the selected
commit. SOPS uses the target's key environment. Retrieval/decryption/validation
failures do not advance the applied revision or permit pruning.

## Runtime profile

The runtime profile is separate from the resource configuration in Git:

```yaml
schema_version: 1
source: git
state: .dome/state.db
git_url: ssh://git@git.example.com/configuration
git_ref: main
git_path: config.yaml
ssh_key: /etc/my-agent/git.key
known_hosts: /etc/my-agent/known_hosts
```

For a local source, use `source: local`, `config: config.yaml` and `state`,
without Git fields. Optional `secrets` refers to a local encrypted file for local
sources or a repository-relative encrypted blob for Git sources. `git_allow_http`
defaults to false; `git_ref` defaults to HEAD, `git_path` to config.yaml and
`state` to .dome/state.db. Unknown fields, unsupported schema versions and invalid
source combinations fail. Profiles contain file references, never key contents. HTTP(S) URL userinfo,
query parameters and fragments are rejected; use public HTTPS or verified SSH.
Initialization requires Linux filesystem support for anonymous temporary files
(`O_TMPFILE`) and hard links, plus access to `/proc/self/fd`. Unsupported storage
returns an error; no alternative publication path is selected. New profile files
use mode 0600 and new workspace directories use 0700; existing directory
permissions are preserved.

`--workspace /path` selects the workspace (default: current directory; no parent
or home search). `--agent-config settings.yaml` selects an explicit profile;
`agent init --agent-config settings.yaml --non-interactive` saves it into the
workspace. Relative host paths in either profile resolve against the selected
workspace, while Git blob paths remain repository-relative. Workspace selection
does not change process working directory or consumer-defined resource paths.

Complete input needs no prompts. `--non-interactive`, `--json` and non-terminal
input never prompt; missing inputs fail. Other commands never prompt. Repeating
init with equivalent settings is a no-op; conflicting settings fail without
replacing the profile. Deliberately edit the profile to change settings, preserving
state location and resource identities. Existing local configuration and keys
are never overwritten. SSH key and known-hosts references must be readable files.
Successful initialization validates settings and local references, not Git
connectivity, remote configuration contents or SOPS decryption.

## Compatibility overrides

Existing Go-configured sources and explicit runtime flags still work without an
initialized profile. Legacy source, credential, secrets and state flags remain
accepted but hidden on runtime/root help; setup flags are listed on `init --help`.
Explicit flags override the selected profile, which overrides Go defaults, without
rewriting it. Explicit local/Git source selection replaces source-specific saved
settings, including secrets and credentials; supply needed values for that new
source. Ref/path-only flags refine the saved Git source. Invalid selected profiles
fail before runtime execution; they never cause silent source fallback.
Direct library calls using `cli.Options` do not discover workspace profiles.
