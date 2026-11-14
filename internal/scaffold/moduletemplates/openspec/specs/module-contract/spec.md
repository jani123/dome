# Module contract

## Purpose

Describe the generated resource module's public contract and explicit unfinished host adapter boundary, without claiming support for an unimplemented host integration.

## Requirements

### Requirement: Safe public resource contract

Resources SHALL validate stable non-secret names, observe actual backend state, apply repeatable actions, and restore deletion handles from identity without old secrets. A missing backend SHALL fail explicitly.

#### Scenario: Backend is unfinished
- **WHEN** observation or mutation is requested without an implemented backend
- **THEN** the operation fails instead of claiming successful convergence

#### Scenario: Prune previously managed resource
- **WHEN** a registered module restores a valid managed identity
- **THEN** it returns a validated deletion handle without old secret values
