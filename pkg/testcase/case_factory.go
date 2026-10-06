package testcase

import (
	"fmt"
	"path/filepath"
	"strings"
)

// LoadTestCases load test cases by file path
// Supported formats: CSV (data in separate sheet files) and JSON (data inline).
func LoadTestCases(filePath string) ([]*TestCase, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	var parser CaseParser

	switch ext {
	case ".csv":
		parser = &CSVCaseParser{FilePath: filePath}
	case ".json":
		parser = &JSONCaseParser{FilePath: filePath}
	// TODO: Excel (.xls/.xlsx) support - needs a spreadsheet dependency
	// (e.g. excelize); see docs/design.md for the extension guide.
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}

	return parser.Parse()
}
