package testcase

import (
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

// Validate expect and actual
func (t *TestStep) Validate() validate.CompareResult {
	result := validate.CompareStruct(t.Expect, t.actual)
	return result
}

// StepValidateResult record validate result for step
type StepValidateResult struct {
	Index  int
	StepID string
	Passed bool
	Detail validate.CompareResult
}

// TestCase represents a test case with its steps.
type TestCase struct {
	CaseNo          string
	CaseTitle       string
	Steps           []TestStep
	ValidateResults []StepValidateResult
}

// AddValidateResult collect validate result for test case
func (t *TestCase) AddValidateResult(index int, stepID string, result validate.CompareResult) {
	t.ValidateResults = append(t.ValidateResults, StepValidateResult{
		Index:  index,
		StepID: stepID,
		Passed: result.Equal,
		Detail: result,
	})
}

// CaseParser is an interface for parsing test cases from different formats.
type CaseParser interface {
	Parse() ([]*TestCase, error)
}
