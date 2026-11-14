# Change review

## Purpose

Provide one review entrypoint that combines independent specialist analysis with actual formatting, static checks, and test evidence for an identified implementation and its OpenSpec plan.

## Requirements

### Requirement: Identified review scope

The review skill SHALL identify the target change, comparison revision, and relevant staged, unstaged, and untracked files before delegation. Ambiguous scope SHALL be clarified. Each reviewer SHALL inspect the same identified input state; changes during review SHALL invalidate affected findings or checks until reviewed again.

#### Scenario: Mixed working tree
- **WHEN** a workspace contains multiple active changes and unrelated edits
- **THEN** the review identifies the selected implementation and reports any unavoidable whole-repository verification scope separately

### Requirement: Independent specialist delegation

One skill invocation SHALL coordinate separate subagents for implementation analysis, architecture, security, verification execution, and plan alignment. It SHALL respect available concurrency limits and report missing specialist results as incomplete. If delegation is unavailable it SHALL report that limitation rather than claim independent review.

#### Scenario: Fewer slots than reviewers
- **WHEN** five specialist roles exceed available agent slots
- **THEN** they run in bounded batches with all five roles represented in the final report

### Requirement: Actual formatting and quality checks

The verification reviewer SHALL run non-mutating Go formatting checks, linters/static analysis, and applicable required tests from repository instructions. It SHALL record tool versions, commands, exit outcomes, and coverage limits. Formatting checks SHALL include relevant new Go files and detect unformatted source without fixing it.

#### Scenario: Unformatted added source
- **WHEN** a relevant new Go file violates Go or repository formatting conventions
- **THEN** review reports the file and failing format check without rewriting it

#### Scenario: Integration dependency missing
- **WHEN** a required suite cannot run because Docker or another real dependency is unavailable
- **THEN** the check is reported blocked with the actual cause and the review does not claim full verification

### Requirement: Requirement and design alignment

The plan reviewer SHALL compare selected OpenSpec requirements, scenarios, design decisions, and task claims with source and test evidence. It SHALL report missing behavior, unsupported completion claims, unplanned behavior, and contradictions without silently changing the plan or checking tasks complete.

#### Scenario: Code diverges from plan
- **WHEN** implementation omits a specified behavior or introduces an incompatible decision
- **THEN** the report identifies the affected artifact and source evidence and recommends resolving the discrepancy explicitly

### Requirement: Evidence-backed specialist findings

Implementation, architecture, and security reviewers SHALL inspect relevant source and boundaries and provide actionable findings with severity, source location, triggering conditions, and impact. The coordinator SHALL reconcile duplicate or contradictory findings using evidence and distinguish demonstrated defects from unverified risks.

#### Scenario: Two reviewers find one defect
- **WHEN** reviewers independently identify the same root cause
- **THEN** the final report contains one reconciled finding retaining the relevant evidence and specialist coverage

### Requirement: Review preserves the work

Review SHALL NOT edit implementation, reformat source, change OpenSpec completion state, commit, publish, or archive. Check-generated logs and caches SHALL remain in ignored or isolated storage. Shared-output verification SHALL have a single execution owner. Required failed or blocked checks SHALL remain visible in the final assessment.

#### Scenario: Review finds a defect
- **WHEN** a reviewer identifies a bug or formatting failure
- **THEN** it returns evidence and a recommendation while leaving the reviewed source and task state unchanged

### Requirement: One complete review result

The coordinator SHALL return one report identifying scope, all five reviewer outcomes, reconciled findings, command results, and missing evidence. It SHALL NOT claim a complete passing review when a required specialist or check failed, was blocked, was cancelled, or became stale after source changes.

#### Scenario: Partial agent completion
- **WHEN** four specialists return and one fails or is interrupted
- **THEN** the report names the missing role and reports review incompleteness without treating silence as success

### Requirement: Formal specialist reports

Every specialist SHALL return a report conforming to one versioned schema with role, reviewed input identity, outcome, coverage, findings, evidence, and limitations. Empty findings SHALL be explicit. Verification reports SHALL include structured command outcomes. Invalid or missing reports SHALL be reported as incomplete, never silently counted as passing.

#### Scenario: Specialist finds no issues
- **WHEN** a specialist completes its assigned review without findings
- **THEN** it returns a valid report with an empty findings list, inspected coverage and evidence, and any limitations

#### Scenario: Invalid specialist report
- **WHEN** a specialist returns free-form text or omits required report fields
- **THEN** the coordinator requests a corrected report and reports the role incomplete if a valid report is not obtained

### Requirement: Traceable consolidated report

The coordinator SHALL produce a structured aggregate linking each role's report, preserving input identity, check outcomes, limitations, and finding provenance. Deduplicated findings SHALL retain source report and finding identifiers. A human-readable summary SHALL agree with the structured aggregate.

#### Scenario: Consolidate overlapping findings
- **WHEN** two valid specialist reports identify the same defect
- **THEN** the consolidated finding retains references to both original findings without losing conflicting or limiting evidence

## Evidence and verification boundary

Implementation: [review skill](../../../.agents/skills/dome-review/SKILL.md), [report validator](../../../internal/review/report.go), and [report tests](../../../internal/review/report_integration_test.go).

Outstanding acceptance and scoped evidence remain in [add-coordinated-review-skill](../../changes/add-coordinated-review-skill/tasks.md).
