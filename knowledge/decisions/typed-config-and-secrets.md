---
type: Decision
title: "Configuration sources"
description: "Go defines topology while typed YAML and SOPS supply values without persisting secrets."
generated: {by: codex/2026-11-13, at: "2026-11-13T07:03:33Z"}
sources:
  - id: source-1
    resource: ../../config/config.go
  - id: source-2
    resource: ../../config/git.go
  - id: source-3
    resource: ../../config/config_test.go
  - id: source-4
    resource: ../../config/git_test.go
  - id: source-5
    resource: ../../config/sops_integration_test.go
  - id: source-6
    resource: ../../tests/integration/git_test.go
---

# Configuration sources

Typed configuration keeps domain choices in consumer Go instead of introducing another infrastructure language. Source retrieval and decryption must finish successfully before reconciliation can make pruning decisions. Git reads committed blobs into memory, requires SSH host verification, and requires deliberate opt-in for plain HTTP. Keys remain on the host and decrypted values stay ephemeral.

## Boundary

Private HTTPS auth, encrypted SSH private keys, submodules, and Git LFS configuration blobs are unsupported. Retrieval fetches history anew; keep repositories small. Blobs over 16 MiB are rejected.

Behavioral contract: [Configuration sources](../../openspec/specs/configuration-sources/spec.md).
