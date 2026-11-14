package review

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReportConsistency(t *testing.T) {
	schema := []byte(`{"type":"object"}`)
	valid := map[string]any{"schema_version": "1", "review_id": "review", "input_id": "input", "report_id": "arch", "role": "architecture", "outcome": "passed", "coverage": map[string]any{"completed": []string{"boundaries"}}, "findings": []any{}, "evidence": []any{map[string]any{"id": "source", "kind": "source", "outcome": "passed"}}, "limitations": []any{}}
	check := func(v map[string]any) error {
		b, err := json.Marshal(v)
		require.NoError(t, err)
		_, err = Validate(schema, b, "review", "input")
		return err
	}
	require.NoError(t, check(valid))
	b, err := json.Marshal(valid)
	require.NoError(t, err)
	_, err = Validate(schema, b, "other", "input")
	require.Error(t, err)
	valid["findings"] = []any{map[string]any{"id": "F1", "evidence_ids": []string{"absent"}}}
	require.Error(t, check(valid))
	valid["findings"] = []any{}
	valid["coverage"] = map[string]any{"completed": []string{"x"}, "unperformed": []any{map[string]any{"check": "test", "reason": "Docker unavailable"}}}
	require.Error(t, check(valid))
	valid["coverage"] = map[string]any{"completed": []string{"x"}}
	valid["roles"] = []any{map[string]any{"role": "security", "report_id": nil, "reference": nil, "outcome": "incomplete"}}
	require.Error(t, check(valid))
	valid["roles"] = []any{}
	valid["evidence"] = []any{map[string]any{"id": "check", "kind": "command", "command": "go test ./...", "cwd": ".", "outcome": "passed", "exit_code": 1, "started_at": "2026-11-13T00:00:00Z", "finished_at": "2026-11-13T00:00:01Z", "tool_versions": []string{"go 1.26"}, "required": true}}
	require.Error(t, check(valid))
}
