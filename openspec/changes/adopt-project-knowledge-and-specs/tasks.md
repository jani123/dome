# Implementation tasks

## 1. Establish the planning boundary

- [x] 1.1 Confirm GoogleCloudPlatform/open-knowledge-format as the selected upstream.
- [x] 1.2 Record the owner-approved OpenSpec exception while retaining the ad hoc plan/handover ban.
- [x] 1.3 Initialize the standard OpenSpec schema and write this proposal, design, deltas, and implementation tasks using pinned local planning tooling.

## 2. Reproducible development tooling

- [ ] 2.1 Add a pinned native OpenSpec installer with official checksum verification and Linux amd64/arm64 selection; keep runtime modules independent.
- [ ] 2.2 Add maintainer development commands and README setup instructions, including telemetry/update-check settings for automation.
- [ ] 2.3 Install the needed version-matched Codex OpenSpec workflows, including verify; inspect diffs and preserve existing rules and consumer templates.

## 3. Project knowledge and conformance

- [ ] 3.1 Define and document the `knowledge/` hierarchy, index conventions, source links, authorship/verification policy, and authority boundaries.
- [ ] 3.2 Implement the small developer validator against pinned Google OKF v0.2 conformance; keep stricter repository policy separately identified.
- [ ] 3.3 Test valid minimal concepts, unknown fields/types, absent optional metadata, malformed front matter, reserved indexes, source/footnote references, and broken local links.
- [ ] 3.4 Populate linked architecture, decision, and operational concepts for every area in the design's migration matrix, including licensing and the no-ad-hoc-plans decision.

## 4. Retrospective capability baseline

- [ ] 4.1 Audit source/tests/docs/history and record the inspected revision and retrospective-import scope.
- [ ] 4.2 Backfill composition, reconciliation, state, playbook, transition, configuration, module, and consumer CLI specs with scenarios and evidence links.
- [ ] 4.3 Backfill initializer, consumer authoring, source-access licensing, documentation, release distribution, and development-environment specs.
- [ ] 4.4 Review every migration-matrix row for coverage, consistency, support limitations, and evidence; retain useful rationale in OKF without copying manuals.
- [ ] 4.5 Reinspect outstanding external acceptance and create active work for genuinely unfinished delivery verification; do not fabricate past proposals or approvals.

## 5. Ongoing workflow and CI

- [ ] 5.1 Add concise maintainer guidance for reading relevant knowledge/specs, proposing changes, implementing, verifying, updating memory/docs, and archiving completed scope.
- [ ] 5.2 Add Make/Task validation commands and dedicated CI checks for the bundle and strict OpenSpec validation.
- [ ] 5.3 Verify that malformed disposable fixtures fail checks and that ordinary handovers/scratch notes remain ignored while structured artifacts are tracked.
- [ ] 5.4 Exercise clean development setup and a fresh agent session: recover a prior decision, locate its source/spec, and prepare a small correctly scoped change without reading historical handovers.
- [ ] 5.5 Run the checks required for the final implementation, reconcile evidence/limitations, and archive this change only after its implemented scope is verified.
