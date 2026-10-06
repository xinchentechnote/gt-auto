package executor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	risk_bin "github.com/xinchentechnote/fin-proto-risk-bin-go/messages"
	"github.com/xinchentechnote/gt-auto/pkg/testcase"
	"github.com/xinchentechnote/gt-auto/pkg/validate"
)

func TestBuildReport(t *testing.T) {
	e := &CaseExecutor{
		startedAt:  time.Date(2026, 10, 7, 10, 0, 0, 0, time.UTC),
		finishedAt: time.Date(2026, 10, 7, 10, 0, 5, 0, time.UTC),
	}
	e.Cases = []*testcase.TestCase{
		{
			CaseID:    "c_pass",
			CaseTitle: "passing case",
			ValidateResults: []testcase.StepValidateResult{
				{StepID: "s1", Passed: true},
			},
		},
		{
			CaseID:    "c_fail",
			CaseTitle: "failing case",
			ValidateResults: []testcase.StepValidateResult{
				{
					StepID: "s2",
					Passed: false,
					Detail: validate.CompareResult{
						Diffs: []validate.Diff{{Path: "ClOrdID", Expect: "want", Actual: "got"}},
					},
				},
				{StepID: "s3", Passed: false, Error: "receive failed: timeout"},
			},
		},
	}

	report := e.BuildReport()

	if report.Summary.TotalCases != 2 || report.Summary.TotalSteps != 3 ||
		report.Summary.PassedSteps != 1 || report.Summary.FailedSteps != 2 {
		t.Fatalf("unexpected summary: %+v", report.Summary)
	}
	if report.DurationMs != 5000 {
		t.Fatalf("expected 5000ms duration, got %d", report.DurationMs)
	}
	if len(report.Cases) != 2 {
		t.Fatalf("expected 2 cases, got %d", len(report.Cases))
	}
	if !report.Cases[0].Passed || len(report.Cases[0].Steps) != 1 {
		t.Fatalf("passing case reported wrong: %+v", report.Cases[0])
	}
	fail := report.Cases[1]
	if fail.Passed {
		t.Fatal("failing case should not pass")
	}
	if fail.Steps[0].Diffs[0].Path != "ClOrdID" || fail.Steps[0].Diffs[0].Expect != "want" {
		t.Fatalf("diff not carried: %+v", fail.Steps[0].Diffs)
	}
	if fail.Steps[1].Error == "" {
		t.Fatal("step error not carried")
	}
}

// TestBuildReportAlwaysMarshallable verifies a report can be encoded even when
// a diff holds an unmarshalable value; the value degrades to a string.
func TestBuildReportAlwaysMarshallable(t *testing.T) {
	e := &CaseExecutor{startedAt: time.Now(), finishedAt: time.Now().Add(time.Second)}
	e.Cases = []*testcase.TestCase{{
		CaseID: "c1",
		ValidateResults: []testcase.StepValidateResult{{
			StepID: "s1",
			Passed: false,
			Detail: validate.CompareResult{Diffs: []validate.Diff{
				{
					Path:   "structured",
					Expect: risk_bin.OrderConfirm{ClOrdId: "want"},
					Actual: risk_bin.OrderConfirm{ClOrdId: "got"},
				},
				{Path: "unmarshalable", Expect: make(chan int), Actual: nil},
			}},
		}},
	}}

	report := e.BuildReport()
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("report must always marshal, got: %v", err)
	}
	text := string(data)
	if !strings.Contains(text, `"path":"structured"`) {
		t.Fatalf("structured diff missing: %s", text)
	}
	if !strings.Contains(text, `"ClOrdID":"got"`) {
		t.Fatalf("expected structured diff payload, got: %s", text)
	}
	if !strings.Contains(text, `"path":"unmarshalable"`) {
		t.Fatalf("unmarshalable diff missing: %s", text)
	}
}

func TestWriteReport(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "report.json")

	report := Report{Summary: RunSummary{TotalCases: 1, FailedSteps: 2}}
	if err := WriteReport(report, path); err != nil {
		t.Fatalf("WriteReport failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read report: %v", err)
	}
	var decoded Report
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("report is not valid JSON: %v", err)
	}
	if decoded.Summary.TotalCases != 1 || decoded.Summary.FailedSteps != 2 {
		t.Fatalf("unexpected decoded summary: %+v", decoded.Summary)
	}
}
