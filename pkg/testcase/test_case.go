package testcase

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/xinchentechnote/gt-auto/pkg/validate"
)

// TestStep represents a single step in a test case.
type TestStep struct {
	StepID        string
	SleepTime     string
	Desc          string
	ActionType    string
	TestTool      string
	TestFunction  string
	TestDataSheet string
	TestData      map[string]any
	Expect        any
	actual        any
}

// SetActual set recieve actual data
func (t *TestStep) SetActual(actual interface{}) {
	t.actual = actual
}

// SetExpect sets the expected value for the step.
func (t *TestStep) SetExpect(expect interface{}) {
	t.Expect = expect
}

// DiffReporter is a custom reporter for cmp.Diff that collects differences.
type DiffReporter struct {
	path  cmp.Path
	diffs []string
}

// PushStep adds a step to the path.
func (r *DiffReporter) PushStep(ps cmp.PathStep) { r.path = append(r.path, ps) }

// PopStep removes the last step from the path.
func (r *DiffReporter) PopStep() { r.path = r.path[:len(r.path)-1] }

// Report is called for each difference found by cmp.Diff.
func (r *DiffReporter) Report(result cmp.Result) {
	if !result.Equal() {
		vx, vy := r.path.Last().Values()
		r.diffs = append(r.diffs, fmt.Sprintf(
			"%-15s | %-8v | %-8v",
			r.path.String(),
			formatValue(vx),
			formatValue(vy),
		))
	}
}

func formatValue(v reflect.Value) interface{} {
	if !v.IsValid() {
		return "<nil>"
	}
	return v.Interface()
}

// Validate expect and actual
func (t *TestStep) Validate() (validate.CompareResult, error) {
	delete(t.TestData, "StepId")
	result, err := validate.CompareJSON(t.TestData, t.actual)
	r := &DiffReporter{}
	cmp.Diff(t.Expect, t.actual, cmp.Reporter(r))
	var b strings.Builder
	if len(r.diffs) > 0 {
		b.WriteString("+---------------+----------+--------+\n")
		b.WriteString("|     PATH      | EXPECTED | ACTUAL |\n")
		b.WriteString("+---------------+----------+--------+\n")
		for _, d := range r.diffs {
			b.WriteString(fmt.Sprintf("| %-13s |\n", d))
		}
		b.WriteString("+---------------+----------+--------+\n")
	}
	result.DiffInfo = b.String()
	return result, err
}

// TestCase represents a test case with its steps.
type TestCase struct {
	CaseNo    string
	CaseTitle string
	Steps     []TestStep
}

// CaseParser is an interface for parsing test cases from different formats.
type CaseParser interface {
	Parse() ([]TestCase, error)
}
