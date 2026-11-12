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

The starter reads `config.yaml` locally. The consumer CLI also supports:

```sh
./agent apply --json --git-url ssh://git@git.example.com/configuration \
  --git-ref main --git-path config.yaml \
  --ssh-key /etc/my-agent/git.key --known-hosts /etc/my-agent/known_hosts
```

Verify SSH host keys through a trusted channel. Public HTTPS retrieval is
supported. Plain HTTP requires explicit `--git-allow-http`; never enable it
implicitly. Private HTTPS authentication, encrypted SSH private keys, submodules,
and LFS configuration blobs are not supported. Keep configuration repositories
small because retrieval fetches history afresh in memory.

With Git, `--secrets secrets.enc.yaml` reads encrypted secrets from the selected
commit. SOPS uses the target's key environment. Retrieval/decryption/validation
failures do not advance the applied revision or permit pruning.
