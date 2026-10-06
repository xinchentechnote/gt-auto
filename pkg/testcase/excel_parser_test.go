package testcase

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

func TestExcelCaseParserParse(t *testing.T) {
	parser := &ExcelCaseParser{FilePath: filepath.Join("testdata", "risk_test_case.xlsx")}
	cases, err := parser.Parse()

	assert.NoError(t, err)
	assert.Len(t, cases, 1)

	tc := cases[0]
	assert.Equal(t, "risk_001", tc.CaseID)
	assert.Equal(t, "order", tc.CaseTitle)
	assert.Len(t, tc.Steps, 4)

	assert.Equal(t, "new_order_001", tc.Steps[0].StepID)
	assert.Equal(t, "100101", tc.Steps[0].MsgType)
	assert.Equal(t, "1", tc.Steps[0].SleepMs)
	assert.Equal(t, "risk_bin_oms_1", tc.Steps[0].TestTool)
	assert.False(t, tc.Steps[0].VerifyRequired)
	assert.True(t, tc.Steps[1].VerifyRequired)
	assert.Equal(t, "c00001", tc.Steps[0].TestDatas["ClOrdID"])
	assert.Equal(t, "200102", tc.Steps[2].MsgType)
	assert.Equal(t, "ORIGCLORD1", tc.Steps[2].TestDatas["ClOrdID"])
	for _, step := range tc.Steps {
		assert.Empty(t, step.SkipReason)
	}
}

// TestExcelCaseParserMissingDataSheet verifies a case row referencing a
// non-existent data sheet keeps the step with a SkipReason.
func TestExcelCaseParserMissingDataSheet(t *testing.T) {
	f := excelize.NewFile()
	rows := [][]any{
		{"case_id", "case_title", "step_id", "sleep_ms", "step_desc", "action_type", "verify_required", "test_tool", "msg_type", "test_data"},
		{"c1", "title", "s1", 1, "d", "Send", "N", "tool", "100101", "no_such_sheet"},
	}
	for i, row := range rows {
		cell, _ := excelize.CoordinatesToCellName(1, i+1)
		assert.NoError(t, f.SetSheetRow(f.GetSheetName(0), cell, &row))
	}
	path := filepath.Join(t.TempDir(), "case.xlsx")
	assert.NoError(t, f.SaveAs(path))

	cases, err := (&ExcelCaseParser{FilePath: path}).Parse()
	assert.NoError(t, err)
	assert.Len(t, cases, 1)
	assert.Len(t, cases[0].Steps, 1, "unresolvable step must not be dropped")
	assert.Nil(t, cases[0].Steps[0].TestDatas)
	assert.Contains(t, cases[0].Steps[0].SkipReason, "no_such_sheet")
}

func TestLoadTestCasesXLSX(t *testing.T) {
	cases, err := LoadTestCases(filepath.Join("testdata", "risk_test_case.xlsx"))
	assert.NoError(t, err)
	assert.Len(t, cases, 1)
	for _, step := range cases[0].Steps {
		assert.NotNil(t, step.TestDatas)
		assert.Empty(t, step.SkipReason)
	}
}
