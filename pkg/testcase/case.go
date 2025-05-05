package testcase

// TestStep represents a single step in a test case.
type TestStep struct {
	StepID        string
	SleepTime     string
	Desc          string
	ActionType    string
	TestTool      string
	TestFunction  string
	TestDataSheet string
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
