package testcase

import (
	"fmt"
	"path/filepath"
	"strings"
)

// LoadTestCases load test cases by file path
// It's just support csv now
func LoadTestCases(filePath string) ([]*TestCase, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	var parser CaseParser

	switch ext {
	case ".csv":
		parser = &CSVCaseParser{FilePath: filePath}
	// TODO
	// case ".json":
	// 	parser = &JSONCaseParser{FilePath: filePath}
	// case ".xls", ".xlsx":
	// 	parser = &ExcelCaseParser{FilePath: filePath}
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}

	return parser.Parse()
}
