# Spec Delta

## ADDED Requirements

### Requirement: Strict workspace bootstrap configuration

Workspace settings SHALL use versioned typed YAML separate from resource configuration. Unknown fields, malformed input, unsupported versions and invalid source combinations SHALL fail before reconciliation or pruning. Profiles SHALL store non-secret source settings and credential-file references, never private keys or decrypted secrets.

#### Scenario: Invalid saved profile

- **WHEN** a selected workspace profile is invalid or unreadable
- **THEN** runtime commands needing that profile fail before backend execution without falling back to another source or advancing applied state

#### Scenario: Separate bootstrap settings from resource values

- **WHEN** an operator supplies bootstrap configuration to init
- **THEN** it configures workspace/source settings while resource values remain subject to the compiled agent's existing typed configuration validation

### Requirement: Deterministic saved settings and explicit overrides

Runtime source/state commands SHALL resolve explicit flags over saved workspace settings over existing Go defaults. Explicit source selection SHALL replace the saved source as a coherent unit. Flag defaults SHALL NOT override saved values unless explicitly supplied. Overrides SHALL NOT rewrite saved profiles. Uninitialized legacy workflows SHALL remain supported.

#### Scenario: Apply using saved Git settings

- **WHEN** apply runs without source flags in an initialized Git workspace
- **THEN** it retrieves the saved committed configuration and secrets with the saved ref and transport verification policy

#### Scenario: Override a saved source

- **WHEN** an invocation explicitly selects a local configuration instead of the saved Git source
- **THEN** it uses that local source without inheriting incompatible Git credential or repository-secret paths, and leaves the profile unchanged

#### Scenario: Explicit false value

- **WHEN** an invocation explicitly disables a saved HTTP opt-in
- **THEN** the explicit false value takes precedence and HTTP retrieval is rejected

#### Scenario: Preserve existing Go consumers

- **WHEN** no workspace profile is selected or present and an agent has an existing Go-configured or explicit-flag source and state path
- **THEN** the invocation retains that behavior without requiring init

### Requirement: Stable workspace paths and verified transport

Saved local configuration, local secret, credential and state paths SHALL resolve relative to the selected workspace unless absolute. Git blob paths SHALL remain repository-relative. Source loading SHALL retain SSH host verification, explicit HTTP opt-in and committed revision coherence. Failed initialization SHALL NOT advance applied state or change resource ownership.

#### Scenario: Run from a different directory

- **WHEN** an operator selects an initialized workspace from another directory
- **THEN** source and state paths resolve identically to use within that workspace without changing process working directory or Go-defined resource paths

#### Scenario: SSH workspace settings

- **WHEN** init selects an SSH source without usable key or known-hosts references
- **THEN** setup cannot complete by disabling host verification or embedding credentials
