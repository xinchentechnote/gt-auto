package testcase

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

// ExcelCaseParser implements the CaseParser interface for Excel workbooks
// (.xlsx). The workbook layout mirrors the CSV format:
//
//   - the FIRST sheet is the case table (same testCaseColumns schema as the
//     CSV main file, header row first);
//   - every other sheet is a test data sheet: header row holds the field
//     names (must include a StepId column), rows are keyed by StepId - the
//     same layout as the CSV data sheet files.
//
// See docs/design.md for the format spec.
type ExcelCaseParser struct {
	FilePath string
}

// Parse parses the workbook and returns the test cases.
func (p *ExcelCaseParser) Parse() ([]*TestCase, error) {
	file, err := excelize.OpenFile(p.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open workbook: %w", err)
	}
	defer func() { _ = file.Close() }()

	sheets := file.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("workbook %s has no sheets", p.FilePath)
	}
	caseSheet := sheets[0]

	// The first sheet is the case table (header row first). Spreadsheet
	// cells are never nil, but excelize trims trailing empty cells, so pad
	// short rows before schema validation.
	var rows [][]string
	sheetRows, err := file.GetRows(caseSheet)
	if err != nil {
		return nil, fmt.Errorf("failed to read case sheet %q: %w", caseSheet, err)
	}
	for _, row := range sheetRows {
		padded := make([]string, testCaseColumns)
		copy(padded, row)
		rows = append(rows, padded)
	}

	return parseCaseRows(rows[1:], func(sheetName string) (map[string]map[string]interface{}, error) {
		if sheetName == caseSheet {
			return nil, fmt.Errorf("sheet %q is the case table, not a test data sheet", sheetName)
		}
		sheetRows, err := file.GetRows(sheetName)
		if err != nil {
			return nil, fmt.Errorf("failed to read sheet %s: %w", sheetName, err)
		}
		return recordsFromRows(sheetRows, fmt.Sprintf("sheet %q in %s", sheetName, p.FilePath))
	})
}
