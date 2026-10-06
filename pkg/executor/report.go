package executor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Report is the structured outcome of a test run, suitable for archiving
// and uploading as a CI artifact.
type Report struct {
	StartedAt  time.Time    `json:"startedAt"`
	FinishedAt time.Time    `json:"finishedAt"`
	DurationMs int64        `json:"durationMs"`
	Summary    RunSummary   `json:"summary"`
	Cases      []CaseReport `json:"cases"`
}

// CaseReport holds the results of one test case. Steps lists only the steps
// that produced a recorded result (a failed Send or a Receive step with
// verification or an execution error); Send steps that succeeded have no
// result and are therefore absent.
type CaseReport struct {
	CaseID    string       `json:"caseId"`
	CaseTitle string       `json:"caseTitle"`
	Passed    bool         `json:"passed"`
	Steps     []StepReport `json:"steps"`
}

// StepReport holds the result of a single step.
type StepReport struct {
	Index  int    `json:"index"`
	StepID string `json:"stepId"`
	Passed bool   `json:"passed"`
	// Error is non-empty when the step could not execute.
	Error string       `json:"error,omitempty"`
	Diffs []DiffReport `json:"diffs,omitempty"`
}

// DiffReport is one field-level difference between expected and actual.
type DiffReport struct {
	Path   string `json:"path"`
	Expect any    `json:"expect,omitempty"`
	Actual any    `json:"actual,omitempty"`
}

// BuildReport assembles the structured report from the executed cases.
func (e *CaseExecutor) BuildReport() Report {
	report := Report{
		StartedAt:  e.startedAt,
		FinishedAt: e.finishedAt,
		DurationMs: e.finishedAt.Sub(e.startedAt).Milliseconds(),
		Summary:    e.summarize(),
		Cases:      make([]CaseReport, 0, len(e.Cases)),
	}
	for _, c := range e.Cases {
		caseReport := CaseReport{
			CaseID:    c.CaseID,
			CaseTitle: c.CaseTitle,
			Passed:    true,
			Steps:     []StepReport{},
		}
		for _, result := range c.ValidateResults {
			if !result.Passed {
				caseReport.Passed = false
			}
			step := StepReport{
				Index:  result.Index,
				StepID: result.StepID,
				Passed: result.Passed,
				Error:  result.Error,
			}
			for _, diff := range result.Detail.Diffs {
				step.Diffs = append(step.Diffs, DiffReport{
					Path:   diff.Path,
					Expect: jsonSafe(diff.Expect),
					Actual: jsonSafe(diff.Actual),
				})
			}
			caseReport.Steps = append(caseReport.Steps, step)
		}
		report.Cases = append(report.Cases, caseReport)
	}
	return report
}

// WriteReport marshals the report as indented JSON and writes it to path,
// creating parent directories as needed.
func WriteReport(report Report, path string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create report directory: %w", err)
		}
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}
	return nil
}

// jsonSafe converts an arbitrary diff value into something JSON-encodable.
// Values that cannot be marshaled degrade to their fmt representation so a
// single odd field cannot break the whole report.
func jsonSafe(v any) any {
	if v == nil {
		return nil
	}
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	var out any
	if err := json.Unmarshal(data, &out); err != nil {
		return fmt.Sprintf("%v", v)
	}
	return out
}
