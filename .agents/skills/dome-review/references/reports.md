# Formal reports

Each specialist writes JSON matching [report.schema.json](report.schema.json).
The coordinator writes JSON matching [aggregate.schema.json](aggregate.schema.json).
Use schema version `1`, assigned review/input IDs, and a unique report ID.
All fields are required; use explicit empty arrays and null only where allowed.

Coverage records actual inspected paths/requirements and performed or unperformed
checks. Findings use local stable IDs, severity, condition, impact, source location,
evidence IDs, recommendation and demonstrated/unverified confidence. Specialists
use empty origins; consolidated findings preserve all original report/finding IDs.
Evidence IDs must resolve in the same report. Source evidence references a path
and line; use null command/times/exit code. Command evidence records exact command,
cwd, tool versions, RFC3339 start/end times, exit code and ignored local log path
as reference. Unexecuted blocked/not-applicable checks use null exit/times and
explain why. Never invent command results or include credentials.

Every role reports `passed`, `findings`, `blocked`, `incomplete`, or `stale`.
Passed requires completed evidence-backed scope, no findings or unperformed work,
and no failed/blocked/cancelled required command. `not_applicable` is a check
outcome with a reason, not a passing execution. Verification includes the selected
command matrix, including blocked checks. Other roles may use source evidence only.

Validate with `go run scripts/check-review-report.go report.json`, adding
`--aggregate` before the filename for a summary. Pass `--review-id` and `--input-id`
to compare with assigned identities. Aggregate validation also loads role reports
relative to its own directory, checks all identities and finding origins, and
rejects paths escaping that review directory. Missing reports are permitted only
for explicitly incomplete roles and cannot yield a passing summary.

Do not trust a structurally valid report without inspecting its evidence. Return
one human-readable summary consistent with the aggregate, with limitations and
failed/blocked work visible. Reports and command logs stay in ignored storage.
