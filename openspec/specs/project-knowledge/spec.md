# Project knowledge

## Purpose

Maintain source-linked project context and verified feature changes through the
Google OKF and OpenSpec workflow introduced by
[adopt-project-knowledge-and-specs](../../changes/adopt-project-knowledge-and-specs/proposal.md).
Implementation evidence: [maintainer guide](../../../docs/development.md),
[validator](../../../internal/knowledge/validate.go),
[validator tests](../../../internal/knowledge/validate_test.go), and
[artifact CI](../../../.github/workflows/ci.yml).

## Requirements

### Requirement: Portable project memory

The repository SHALL maintain curated long-term knowledge under `knowledge/` in
Google Open Knowledge Format v0.2, using readable Markdown with YAML front matter
and index-based selective discovery. Reading the bundle SHALL NOT require a
hosted memory service, a third-party OKF engine, or Google Cloud credentials.

#### Scenario: Recover relevant context

- **WHEN** a new maintainer or agent needs the rationale for a Dome decision
- **THEN** they can locate it through the bundle index and read the concept and its linked sources without loading a chat transcript

### Requirement: Traceable claims and honest verification

Knowledge concepts SHALL identify their sources and generation provenance under
the repository's authoring policy. Verification metadata SHALL describe checks
that actually occurred and SHALL NOT imply human review, current passing tests,
or public deployment without supporting evidence.

#### Scenario: Import historical test evidence

- **WHEN** an agent imports an earlier report of passing tests
- **THEN** it preserves the historical scope and does not mark the current checkout as verified by that report

#### Scenario: Describe an unpublished release

- **WHEN** binary packaging has been exercised only locally
- **THEN** the knowledge distinguishes that evidence from public tagged-module, native-platform, and deployment acceptance

### Requirement: Canonical information boundaries

The knowledge bundle SHALL preserve durable rationale and source-linked facts
while referencing canonical requirements and usage documentation. It SHALL NOT
store raw conversations, ad hoc plans, session progress logs, or secrets as
project memory.

#### Scenario: Preserve a customer-ownership decision

- **WHEN** the licensing decision is imported
- **THEN** the concept explains its customer-independence rationale and links the license and licensing documentation rather than duplicating their full text

### Requirement: Conformance distinct from repository policy

Developer validation SHALL distinguish Google OKF v0.2 structural conformance
from stricter Dome authoring requirements. It SHALL accept unknown concept types
and extension fields and SHALL NOT treat optional upstream metadata as required
for format conformance. Dome policy SHALL additionally check its chosen evidence
and local-reference conventions.

#### Scenario: Validate an extensible concept

- **WHEN** an otherwise conforming concept has an unknown type or extension key
- **THEN** the format validator accepts it

#### Scenario: Enforce repository link quality

- **WHEN** a maintained concept has a broken local evidence reference
- **THEN** repository-policy validation reports it without mislabeling it an upstream format violation
