//go:build integration

package review

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestActualReportSchema(t *testing.T) {
	schema, err := os.ReadFile("../../.agents/skills/dome-review/references/report.schema.json")
	require.NoError(t, err)
	evidence := map[string]any{"id": "E1", "kind": "source", "reference": "resource.go:1", "command": nil, "cwd": nil, "tool_versions": []string{}, "started_at": nil, "finished_at": nil, "exit_code": nil, "outcome": "passed", "required": true, "reason": nil}
	report := map[string]any{"schema_version": "1", "review_id": "R1", "report_id": "A1", "input_id": "I1", "role": "architecture", "outcome": "passed", "coverage": map[string]any{"inspected": []string{"resource.go"}, "completed": []string{"public boundaries"}, "unperformed": []any{}}, "findings": []any{}, "evidence": []any{evidence}, "limitations": []any{}}
	validate := func(v map[string]any) error {
		b, err := json.Marshal(v)
		require.NoError(t, err)
		_, err = Validate(schema, b, "R1", "I1")
		return err
	}
	require.NoError(t, validate(report))
	report["role"] = "unknown"
	require.Error(t, validate(report))
	report["role"] = "architecture"
	delete(report, "limitations")
	require.Error(t, validate(report))
	report["limitations"] = []any{}
	report["outcome"] = "blocked"
	evidence["kind"] = "command"
	evidence["command"] = "go test ./..."
	evidence["cwd"] = "."
	evidence["outcome"] = "blocked"
	evidence["reason"] = "Docker unavailable"
	require.NoError(t, validate(report))
	report["outcome"] = "passed"
	require.Error(t, validate(report))
	report["outcome"] = "blocked"
	evidence["reason"] = nil
	require.Error(t, validate(report))
	report["outcome"] = "incomplete"
	evidence["reason"] = "not executed"
	report["role"] = "verification"
	require.NoError(t, validate(report))
	report["evidence"] = []any{}
	require.Error(t, validate(report))
	aggregate, err := os.ReadFile("../../.agents/skills/dome-review/references/aggregate.schema.json")
	require.NoError(t, err)
	delete(report, "role")
	report["outcome"] = "incomplete"
	report["roles"] = []any{}
	b, err := json.Marshal(report)
	require.NoError(t, err)
	_, err = Validate(aggregate, b, "R1", "I1")
	require.Error(t, err)
	roles := []any{}
	for _, role := range []string{"implementation", "architecture", "security", "verification", "plan"} {
		roles = append(roles, map[string]any{"role": role, "report_id": nil, "reference": nil, "outcome": "incomplete"})
	}
	report["roles"] = roles
	b, err = json.Marshal(report)
	require.NoError(t, err)
	_, err = Validate(aggregate, b, "R1", "I1")
	require.NoError(t, err)
	report["outcome"] = "passed"
	b, err = json.Marshal(report)
	require.NoError(t, err)
	_, err = Validate(aggregate, b, "R1", "I1")
	require.Error(t, err)
}
