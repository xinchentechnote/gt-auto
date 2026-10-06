package testcase

import (
	"fmt"
	"path/filepath"
	"strings"
)

// LoadTestCases load test cases by file path
// Supported formats: CSV and Excel (data in separate sheets), JSON (data inline).
func LoadTestCases(filePath string) ([]*TestCase, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	var parser CaseParser

	switch ext {
	case ".csv":
		parser = &CSVCaseParser{FilePath: filePath}
	case ".json":
		parser = &JSONCaseParser{FilePath: filePath}
	case ".xlsx", ".xlsm":
		parser = &ExcelCaseParser{FilePath: filePath}
	// Legacy .xls is not supported by excelize; convert to .xlsx first.
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}

	return parser.Parse()
}
