# Verification: project knowledge and specification workflow

## Scope

Local implementation is verified against the proposal, deltas, design, and
migration inventory. The change remains active because a real fresh maintainer
agent session has not been exercised. A deterministic clean-checkout smoke is
recorded separately below and is not a substitute for that acceptance.

## Completeness

The repository contains 14 retrospectively imported capability specs, 15 curated
OKF concepts, and two new workflow capability specs. The coverage matrix in
[the maintainer guide](../../../docs/development.md) maps all existing areas to
knowledge and implementation/test evidence. Their baseline revision is explicitly
identified; no historical proposals, approvals, or passing runs were fabricated.

The pinned native installer, isolated wrapper, seven generated workflows,
upstream MIT attribution, developer commands, Devbox setup, artifact CI, and
validator are implemented. The maintainer rules and consumer skill template were
byte-checked before/after repeatable workflow generation.

Fresh-agent acceptance (task 5.4) is **not verified**. Archiving (5.6) is deferred
until that check is performed. The separate `verify-public-delivery` change also
tracks live consumer Codex behavior, native arm64/public bootstrap, indexing,
GitHub settings, and deployment acceptance.

## Correctness and scenario evidence

| Requirement area | Evidence |
| --- | --- |
| Portable knowledge and selective discovery | `knowledge/index.md` plus directory indexes and local source/spec links; clean-checkout retrieval smoke |
| Honest provenance | Generated actors/times and source references; no invented human verification; timestamp-offset and verification-order tests |
| Canonical roles | `AGENTS.md`, maintainer guide, source-linked concepts, and capability specs; no committed handovers |
| Format versus repository policy | Minimal/unknown-type/extension fixtures pass conformance while missing evidence and bad links fail policy |
| Existing-work import | Maintainer coverage matrix and 14 baseline specs with implementation/tests, explicit retrospective scope, and support limitations |
| Unfinished work | Active `verify-public-delivery` with actual dated external observations and unchecked acceptance |
| Pinned maintainer workflows | Checksummed OpenSpec 2.0.0 installation; generator metadata drift tests and `check-specs` gate |
| Runtime independence | Knowledge tooling is imported only by developer checks; no consumer runtime or Go dependency versions changed |

Unit tests use an injected in-memory filesystem. Real filesystem/symlink checks
are integration-tagged and use isolated temporary roots. Diagnostics do not echo
malformed document contents. Fenced/inline examples do not become false links or
claims; parent indexes, descriptions, source IDs, local references, and timestamp
text are checked. This is not a full Markdown renderer, external freshness check,
or proof of narrative truth; those limits are documented.

## Checks observed locally

On 2026-11-13, `make check`, `make test-race`, `make test-e2e`, and
`make test-integration` passed. The real Git HTTP/SSH suite and seven compiled-agent
container E2E scenarios ran without skips. Subsequent validator refinements passed
their unit, race, and real-filesystem integration checks. Artifact validation,
Hugo/link builds, actionlint, shell syntax, and staged whitespace checks passed.

An isolated clean Git checkout in the existing devcontainer ran `task dev:setup`,
checksum-installed the pinned tool, validated knowledge/specs, and refreshed the
seven maintainer workflows without tracked modifications. A separate deterministic
process read the indexes and customer-independence concept, looked up the matching
spec, and created/validated a documentation-only proposal. Historical handover and
plan files were absent. This was **not** a rebuilt devcontainer, a native arm64 run,
or a live coding-agent session.

A disposable minimal OKF concept passed format-only checks and failed Dome policy
as intended. Ignore checks kept ordinary plans/handovers local while allowing
structured OpenSpec artifacts. Full logs and disposable fixtures remain outside
version control; this record preserves the verification scope and limitations.

## Coherence and disposition

The implementation follows Google's format rather than a third-party memory
engine. Maintainer artifacts remain separate from public usage docs and generated
consumer skills. The no-ad-hoc-plans rule remains in place with the approved
structured OpenSpec exception. Checks are developer-only, and original runtime
behavior is unchanged.

No implementation gap was identified in the checked scope. The applicable live
fresh-agent scenario is not verified, so this report does not claim archive
readiness. Complete task 5.4 and then archive through the normal workflow.
