// Package review validates local developer review evidence, not agent runtime data.
package review

import (
	"encoding/json"
	"fmt"

	"github.com/jani123/dome/core/fault"
	"github.com/xeipuuv/gojsonschema"
)

type (
	Finding struct {
		ID          string   `json:"id"`
		EvidenceIDs []string `json:"evidence_ids"`
		Origins     []Origin `json:"origins"`
	}
	Origin struct {
		ReportID  string `json:"report_id"`
		FindingID string `json:"finding_id"`
	}
	Evidence struct {
		ID           string   `json:"id"`
		Kind         string   `json:"kind"`
		Outcome      string   `json:"outcome"`
		Required     bool     `json:"required"`
		ExitCode     *int     `json:"exit_code"`
		Command      *string  `json:"command"`
		CWD          *string  `json:"cwd"`
		StartedAt    *string  `json:"started_at"`
		FinishedAt   *string  `json:"finished_at"`
		ToolVersions []string `json:"tool_versions"`
		Reason       *string  `json:"reason"`
	}
	Role struct {
		Role      string  `json:"role"`
		ReportID  *string `json:"report_id"`
		Reference *string `json:"reference"`
		Outcome   string  `json:"outcome"`
	}
	Report struct {
		SchemaVersion string `json:"schema_version"`
		ReviewID      string `json:"review_id"`
		ReportID      string `json:"report_id"`
		InputID       string `json:"input_id"`
		Role          string `json:"role"`
		Outcome       string `json:"outcome"`
		Roles         []Role `json:"roles"`
		Coverage      struct {
			Inspected   []string `json:"inspected"`
			Completed   []string `json:"completed"`
			Unperformed []struct {
				Check  string `json:"check"`
				Reason string `json:"reason"`
			} `json:"unperformed"`
		} `json:"coverage"`
		Findings    []Finding  `json:"findings"`
		Evidence    []Evidence `json:"evidence"`
		Limitations []string   `json:"limitations"`
	}
)

// Validate checks schema and cross-field consistency. Schemas are trusted repository assets.
func Validate(schema, body []byte, reviewID, inputID string) (r Report, err error) {
	defer func() { err = fault.Wrap(fault.Validation, err) }()
	result, err := gojsonschema.Validate(gojsonschema.NewBytesLoader(schema), gojsonschema.NewBytesLoader(body))
	if err != nil {
		return r, fmt.Errorf("invalid schema or report: %w", err)
	}
	if !result.Valid() {
		return r, fmt.Errorf("report does not match schema (%d validation errors)", len(result.Errors()))
	}
	if err = json.Unmarshal(body, &r); err != nil {
		return r, err
	}
	if reviewID != "" && r.ReviewID != reviewID {
		return r, fmt.Errorf("review identity mismatch")
	}
	if inputID != "" && r.InputID != inputID {
		return r, fmt.Errorf("input identity mismatch")
	}
	commandCount := 0
	evidence := map[string]bool{}
	failed := false
	for _, e := range r.Evidence {
		if evidence[e.ID] {
			return r, fmt.Errorf("duplicate evidence id")
		}
		evidence[e.ID] = true
		if e.Required && e.Outcome != "passed" && e.Outcome != "not_applicable" {
			failed = true
		}
		if e.Kind == "source" && (e.Command != nil || e.CWD != nil || e.ExitCode != nil || e.StartedAt != nil || e.FinishedAt != nil) {
			return r, fmt.Errorf("source evidence cannot claim command execution")
		}
		if e.Kind == "command" {
			commandCount++
			if e.Command == nil || e.CWD == nil {
				return r, fmt.Errorf("command evidence requires command and cwd")
			}
			executed := e.Outcome == "passed" || e.Outcome == "failed"
			if executed && (e.ExitCode == nil || e.StartedAt == nil || e.FinishedAt == nil || len(e.ToolVersions) == 0) {
				return r, fmt.Errorf("executed command lacks result evidence")
			}
			if e.Outcome == "passed" && *e.ExitCode != 0 {
				return r, fmt.Errorf("passing command has nonzero exit")
			}
			if e.Outcome == "failed" && *e.ExitCode == 0 {
				return r, fmt.Errorf("failed command has zero exit")
			}
			if !executed && (e.Reason == nil || *e.Reason == "") {
				return r, fmt.Errorf("unexecuted check requires reason")
			}
			if !executed && e.ExitCode != nil {
				return r, fmt.Errorf("unexecuted check cannot claim exit code")
			}
			if e.Required && e.Outcome != "passed" && e.Outcome != "not_applicable" {
				failed = true
			}
		}
	}
	findings := map[string]bool{}
	for _, f := range r.Findings {
		if findings[f.ID] {
			return r, fmt.Errorf("duplicate finding id")
		}
		findings[f.ID] = true
		for _, id := range f.EvidenceIDs {
			if !evidence[id] {
				return r, fmt.Errorf("finding references missing evidence")
			}
		}
	}
	if r.Role == "verification" && commandCount == 0 {
		return r, fmt.Errorf("verification report requires command matrix")
	}
	if r.Outcome == "passed" && (len(r.Findings) > 0 || len(r.Coverage.Unperformed) > 0 || failed || len(r.Coverage.Completed) == 0 || len(r.Evidence) == 0) {
		return r, fmt.Errorf("unsupported passing outcome")
	}
	if r.Outcome == "findings" && len(r.Findings) == 0 {
		return r, fmt.Errorf("findings outcome requires findings")
	}
	seen := map[string]bool{}
	for _, role := range r.Roles {
		if seen[role.Role] {
			return r, fmt.Errorf("duplicate role")
		}
		seen[role.Role] = true
		if role.ReportID == nil || role.Reference == nil {
			if role.Outcome != "incomplete" {
				return r, fmt.Errorf("missing role report must be incomplete")
			}
		}
		if r.Outcome == "passed" && role.Outcome != "passed" {
			return r, fmt.Errorf("aggregate cannot pass with nonpassing role")
		}
	}
	return r, nil
}
