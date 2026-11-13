# Design

## Context

The owner selected GoogleCloudPlatform/open-knowledge-format and explicitly
approved committing OpenSpec artifacts while continuing to reject ad hoc plans,
handovers, and session notes. Existing implementation spans the core library,
recovery/state machinery, consumer tooling, GPLv3 licensing, documentation,
devcontainer networking, and GitHub workflows. Some delivery checks remain
unverified outside the local environment.

Google's upstream is a format specification plus a data-catalog reference agent.
A project knowledge bundle does not need that agent or its BigQuery/Gemini setup.
The inspected OKF v0.2 source is commit
`ad30107c31c06aec8a7d5636e0d1058118604e6f` in the selected repository.

OpenSpec 2.0.0 is the inspected planning tool. Its native Linux x64 release was
checksum-verified and used locally for this proposal. The current container has
Node 20.19.2; the npm distribution requires Node >=22.12.0. Prefer the pinned
native distribution rather than changing Node solely for this workflow.

## Goals and non-goals

Goals: recover durable context across agent sessions; plan changes consistently;
import existing behavior and decisions without losing their evidence; validate
maintained artifacts and prevent contradictory copies.

Non-goals: runtime memory for deployed agents, replacing Go-defined topology,
changing reconciliation semantics, importing raw chats, an external memory
service, automatic GitHub publication, or making these tools mandatory for users
of the Dome library.

## Decisions

### 1. Separate authority by purpose

| Material | Canonical home | Maintenance |
| --- | --- | --- |
| Maintainer rules and non-negotiable invariants | `AGENTS.md` | Explicit repository guidance; never overwritten by a bootstrap tool |
| Supported behavior and acceptance scenarios | `openspec/specs/<capability>/spec.md` | Baseline plus verified deltas |
| Proposed scope, design, implementation and checks | `openspec/changes/<change>/` | Active until complete, then archive |
| Rationale, source-linked facts and operating constraints | `knowledge/` | Curated OKF concepts, reviewed with relevant changes |
| Public usage and development commands | README and existing guides | Reader-oriented documentation |
| Scratch work, logs, generated indexes/caches | Ignored local storage | Never a source of durable authority |

OKF references specs and operational docs instead of copying their requirements
or command manuals. Specs link relevant knowledge to explain a decision. Code and
tests determine what is implemented; a discrepancy becomes explicit active work,
not a silent rewrite of the baseline. User directions and repository rules retain
precedence over generated tool guidance.

### 2. Keep Google's OKF portable and lightweight

Use `knowledge/index.md` with `okf_version: "0.2"` and directory indexes for
selective discovery. Group concepts under `architecture/`, `decisions/`, and
`operations/`. Start with concise, useful concepts rather than one file per
function or a copy of every README section. Use standard Markdown and YAML, with
`type`, title/description, provenance, and links as appropriate.

Use the upstream `generated`, `verified`, `sources`, and lifecycle contracts.
Record real authorship/verification timestamps. Do not claim human review just
because an agent wrote a concept, and do not turn a historical test report into
a current successful result. Add current-state verification only when supported
by source inspection or a relevant check, with the scope stated. Keep source
paths and inspected revisions traceable; don't invent publicly reachable commit
URLs for commits that have not been pushed.

Do not install either `okf-memory/okf-agent-memory` or `serradura/okf`. The selected
Google upstream does not supply those projects' `okf` CLI commands. Implement a
small repository-owned developer validator using existing Go/YAML dependencies,
with a clear distinction between upstream conformance and Dome's authoring rules.
Conformance accepts unknown fields/types and missing optional metadata; the local
authoring policy additionally checks our chosen provenance and local links.
No reference-agent installation, cloud connection, or full-text/vector service
is needed. Initially use indexes and `rg` to retrieve relevant concepts.

### 3. Use OpenSpec's standard repository workflow

Keep the `spec-driven` schema and short project configuration. Install only
maintainer integrations needed here, initially Codex, and include verification
in the selected workflow set. Inspect generated files before accepting them;
retain existing `AGENTS.md`, consumer-skill templates, and unrelated settings.
Use OpenSpec 2.0 workflow names rather than assuming older `/opsx:*` commands.

Pin OpenSpec and maintain the pin in documented developer tooling. Download the
matching official native release for amd64/arm64 and verify published checksums.
Use the same version in CI; disable telemetry and unsolicited update checks in
reproducible automation. Generated skills must match the selected tool version.
The initial binary under `.build/tools/` is local planning tooling, not yet the
maintained installation mechanism.

### 4. Import completed work as a retrospective baseline

Inventory tracked source, tests, documented decisions, and history before writing
baseline specs. Use ignored historical plans only as discovery aids; validate each
retained claim against canonical evidence. Do not resurrect them in the bundle.
Preserve durable outcomes and rationale, rather than a transcript or every edit.

| Existing area | Baseline capability boundary | OKF coverage and primary evidence |
| --- | --- | --- |
| Go composition and public API/examples | `agent-composition` | Explicit modules, external contracts; `agent/`, registry/resource packages, examples and external-consumer tests |
| Observe/decide/apply/verify/persist and pruning | `reconciliation` | Ownership and revision invariants; reconciler and tests |
| SQLite migrations, guards, managed/applied state | `durable-state` | Persistence is not observation; `core/state/` and integration tests |
| DAG steps, locks, retries and recovery branches | `playbook-execution` | Repeatability, failures, graph versions and interrupted/completed-step re-observation; DAG/playbook code and tests |
| Resource transition and interrupted pruning recovery | `resource-transitions` | Pending target protection and verification; reconcile/transition tests |
| Typed YAML, SOPS and local/Git sources | `configuration-sources` | Secret handling, SSH verification, HTTP opt-in and unsupported modes; config code/tests |
| File and systemd resource behavior | `resource-modules` | Atomic files, permissions and service limits; module code/tests; distinguish mocked systemd checks from live coverage |
| Consumer CLI and execution diagnostics | `consumer-cli` | Exit categories, JSON planning, polling, status/resume; CLI/agent code/tests |
| Initializer and version-matched sandbox starter | `project-initialization` | Staging, collision/symlink refusal, cleanup and version guard; scaffold/CLI code and real consumer tests |
| Consumer Codex skill and update handover | `consumer-authoring` | Version checks and user-requested apply; embedded skill and generated instructions; live authoring remains unverified |
| GPLv3 and customer source/build access | `customer-source-access` | License decision, combined-agent boundary and customer independence; LICENSE, licensing guide and dependency review |
| Hugo, guide versions, GitHub Pages | `documentation-publishing` | Pinned tool, snapshot/alias behavior and owner setup; scripts, layouts and workflow |
| Shared version, artifacts and corresponding source | `release-distribution` | RC/stable policy, native gates, checksums and vendored source; scripts/workflows and release docs |
| Devcontainer egress and test infrastructure | `development-environment` | HTTPS proxy versus SSH, allowed hosts, Docker/Testcontainers and Git fixtures; devcontainer, Makefile/Taskfile and tests |

Use cohesive sections within these boundaries, splitting only when requirements
justify it. Each baseline requirement links to an implementation and test, or
explicitly records the absence/limitation of evidence. Add a concise retrospective
marker and inspected revision in the import documentation; do not assert that the
spec was authored before the existing code.

### 5. Keep unfinished acceptance active

Create an active delivery-verification change for the actual remaining scope:
real native arm64 execution, real Codex discovery/authoring, public exact-tag
bootstrap, pkg.go.dev recognition, GitHub settings/required checks, and deployed
Pages/release verification. Recheck current external state rather than importing
old token availability, HTTP responses, or access failures as permanent facts.
Do not reopen already verified library implementation as fictional unfinished
work, and do not archive release acceptance merely because local tests passed.
The Python skill validator's missing dependency is a past local limitation, not
proof of a feature defect or a required new runtime dependency.

### 6. Validate and maintain artifacts with changes

Add documented Make/Task targets and a dedicated CI job for OKF structure/local
policy plus `openspec validate --all --strict`. Changes to commands/developer setup
update README. Changes to capability behavior update specs; changed rationale or
facts update relevant knowledge in the same review. Agents consult the relevant
indexes/specs before implementation and update them after verification.

Keep secrets, private credentials, raw process logs, external caches, and scratch
notes out of tracked bundles. Omit an OKF session-style log; Git already records
artifact history. If a later knowledge log is introduced, it must describe curated
concept changes, not agent session narratives.

## Rollout

First adopt the explicit artifact exception and this structured proposal. Then
add reproducible tooling and knowledge conventions; backfill and review existing
capabilities; record remaining delivery work; add validation and perform a fresh
session exercise. Keep these steps reviewable without pausing merely at a phase
boundary once implementation is authorized. Baseline import can be reviewed in
cohesive groups instead of one large opaque dump.

## Verification

Use unit tests for the validator with malformed YAML, reserved-file conventions,
unknown types/extensions, optional trust metadata, links and provenance cases.
Real filesystem/tool invocation tests use the integration tag. Check failure
behavior by breaking a disposable fixture, not the maintained bundle. Verify a
fresh devcontainer can install pinned OpenSpec and validate the repository, and
that an agent can recover context and propose a small change using both artifacts.
Run repository checks required by AGENTS.md for affected scope; writing this
proposal itself changes no runtime behavior.

## Risks and open implementation details

The main risks are contradictory copies, false verification claims, overly broad
baseline specs, generated-rule overwrites, and tool-version drift. The authority
map, source-linked import, explicit remaining work, narrow integration, and pinned
validation address them. Exact concept/file counts and a possible later graph
viewer are intentionally not fixed now. A viewer and optional search service are
outside initial scope.

## References

- [Selected Google OKF repository](https://github.com/GoogleCloudPlatform/open-knowledge-format)
- [Inspected OKF v0.2 specification](https://github.com/GoogleCloudPlatform/open-knowledge-format/blob/ad30107c31c06aec8a7d5636e0d1058118604e6f/SPEC.md)
- [OpenSpec project setup](https://openspec.dev/docs/setup)
- [OpenSpec project configuration](https://openspec.dev/docs/project-config)
- [OpenSpec installation](https://openspec.dev/docs/installation)
