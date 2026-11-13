# Configuration sources

## Purpose

Retrospective specification of behavior inspected at `3e262920bbc20e41f843250f78a90c4935b9a036`.
This baseline documents existing implementation; it does not assert that these
requirements were authored before the code or that external acceptance passed.

## Requirements

### Requirement: Typed strict configuration

Configuration values SHALL decode into user-defined settings and separate secret structs; unknown YAML fields or malformed input SHALL fail before reconciliation.

Evidence: [config/config_test.go](../../../config/config_test.go).

#### Scenario: Typed strict configuration

- **WHEN** the configuration includes an unknown field
- **THEN** evaluation fails without pruning previously managed resources

### Requirement: Host SOPS decryption

SOPS SHALL use the normal host key environment and SHALL NOT expose decrypted contents in error text, snapshots, or execution metadata.

Evidence: [config/sops_integration_test.go](../../../config/sops_integration_test.go).

#### Scenario: Host SOPS decryption

- **WHEN** the supplied host key cannot decrypt the document
- **THEN** the invocation fails with concealed configuration diagnostics

### Requirement: Verified Git transport

Git sources SHALL support public HTTPS and SSH with private-key authentication plus verified known_hosts; plain HTTP SHALL require explicit opt-in.

Evidence: [config/git_test.go](../../../config/git_test.go).

#### Scenario: Verified Git transport

- **WHEN** the caller uses HTTP without AllowHTTP or --git-allow-http
- **THEN** source validation rejects it

### Requirement: Committed revision coherence

Configuration and encrypted secrets from Git SHALL be read from the selected commit, and failed retrieval/evaluation SHALL prevent reconciliation and revision advancement.

Evidence: [tests/integration/git_test.go](../../../tests/integration/git_test.go).

#### Scenario: Committed revision coherence

- **WHEN** configuration retrieval or decryption fails
- **THEN** the last successful applied revision and managed resources remain protected

### Requirement: Explicit committed-reference retrieval

Git retrieval SHALL resolve an explicitly selected reachable branch, tag, or
commit without requiring a valid remote default HEAD when an explicit branch is
selected.

Evidence: [Git retrieval](../../../config/git.go) and
[real Git transport suite](../../../tests/integration/git_test.go).

#### Scenario: The server default branch is unborn

- **WHEN** a valid committed branch is explicitly selected
- **THEN** configuration retrieval can succeed independently of the unborn default HEAD

## Evidence and limitations

- [config/config.go](../../../config/config.go)
- [config/git.go](../../../config/git.go)
- [config/config_test.go](../../../config/config_test.go)
- [config/git_test.go](../../../config/git_test.go)
- [config/sops_integration_test.go](../../../config/sops_integration_test.go)
- [tests/integration/git_test.go](../../../tests/integration/git_test.go)

Private HTTPS auth, encrypted SSH private keys, submodules, and Git LFS configuration blobs are unsupported. Retrieval fetches history anew; keep repositories small. Blobs over 16 MiB are rejected.

Rationale: [Go defines topology while typed YAML and SOPS supply values without persisting secrets.](../../../knowledge/decisions/typed-config-and-secrets.md).
