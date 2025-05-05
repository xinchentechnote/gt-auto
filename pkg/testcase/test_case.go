package testcase

import "github.com/xinchentechnote/gt-auto/pkg/validate"

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
	actual        any
}

// SetActual set recieve actual data
func (t *TestStep) SetActual(actual interface{}) {
	t.actual = actual
}

// Validate expect and actual
func (t *TestStep) Validate() (validate.CompareResult, error) {
	return validate.CompareJSON(t.TestData, t.actual)
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
