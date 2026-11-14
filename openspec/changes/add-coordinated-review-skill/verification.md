# Review workflow verification

## Scope

The repository-owned skill, formal report/aggregate schemas, deterministic
validator and direct formatting/static/security/workflow tasks are implemented.
Upstream OpenSpec refresh preserved all custom skill files byte-for-byte.
Skill-creator structural validation passed after installing its real PyYAML
dependency through the configured Debian proxy. Schema/semantic tests cover valid
empty findings, missing fields/roles, unsupported passing states, evidence and
input identity. Integration checks exercise real formatting with untracked paths
containing spaces, unchanged source, and a missing formatter.

## Independent forward tests

An independent coordinator used the skill in three isolated repositories, each
with five separate specialist agents scheduled within available slots. A flawed
fixture exposed arithmetic, shell interpretation, global registration, formatting
and unsupported task-completion defects. A second improved fixture correctly
retained injectable-boundary/test-coverage findings instead of manufacturing a
clean pass. A corrected fixture produced explicit empty findings from all four
analysis roles. Actual formatting/vet/unit, race, Staticcheck, artifact and selected
OpenSpec checks passed there. Reports and aggregates validated, deduplication
retained origins, and content manifests confirmed unchanged source/task state.

On report copies, malformed JSON was rejected and a single correction validated;
a missing/interrupted role yielded an incomplete aggregate; mismatched input
identity was rejected and represented as stale; a fabricated passing aggregate
with a missing role was rejected. Original reports remained unchanged.

## Historical proxy blocker (2026-11-13)

At that time, the outer proxy rejected https://vuln.go.dev/index/modules.json.gz
with Forbidden. A scan could not be certified clean. The maintained allowlist
had been updated, but the active environment had not yet been rebuilt.
An updated disposable proxy could not independently egress from the nested
network, and its retry was cancelled rather than counted passing. It was removed.
Reports/logs remain ignored local evidence; no public publishing occurred.

## Scan after environment rebuild (2026-11-14)

Both the official vulnerability database and Docker CE signing-key endpoints
returned HTTP 200 after the user rebuilt the outer environment.
`task check:security` now executes govulncheck 1.9.0 with Go 1.26.0 and reaches
https://vuln.go.dev, but exits nonzero (govulncheck status 3) on reported findings.
The output reports 41 symbol-level vulnerabilities across the Go standard
library, google.golang.org/grpc v1.82.1, and go.opentelemetry.io/otel/sdk v1.44.0.
It also reports additional imported/required vulnerabilities without observed
call paths. These are scanner results, not a manual exploit confirmation.

The network blocker is resolved; the security gate is failing, not blocked or
passing. Dependency/toolchain remediation is separate work; no dependencies were
changed or scanner findings suppressed during this retry. Task 3.4 remains open.
Raw scanner output remains in ignored local storage.
