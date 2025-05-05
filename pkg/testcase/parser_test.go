package testcase

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCSVCaseParserParse(t *testing.T) {
	filePath := filepath.Join("testdata", "test_case.csv")

	parser := &CSVCaseParser{FilePath: filePath}
	cases, err := parser.Parse()

	assert.NoError(t, err)
	assert.Len(t, cases, 1, "should parse 1 test case")

	tc := cases[0]
	assert.Equal(t, "sz_001", tc.CaseNo)
	assert.Equal(t, "order", tc.CaseTitle)
	assert.Len(t, tc.Steps, 4, "should have 4 steps")

	assert.Equal(t, "new_order_001", tc.Steps[0].StepID)
	assert.Equal(t, "szse_bin_oms_1", tc.Steps[0].TestTool)
	assert.Equal(t, "100101", tc.Steps[0].TestFunction)
	assert.Equal(t, "new_order_001", tc.Steps[0].TestData["StepId"])
	assert.Equal(t, "ORDER00001", tc.Steps[0].TestData["ClOrdID"])

	assert.Equal(t, "new_order_002", tc.Steps[1].StepID)
	assert.Equal(t, "szse_bin_tgw_1", tc.Steps[1].TestTool)
	assert.Equal(t, "100101", tc.Steps[1].TestFunction)

	assert.Equal(t, "new_order_003", tc.Steps[2].StepID)
	assert.Equal(t, "szse_bin_tgw_1", tc.Steps[2].TestTool)
	assert.Equal(t, "200102", tc.Steps[2].TestFunction)
}

func TestLoadCSVToMap(t *testing.T) {
	data, err := LoadCSVToMap("testdata/szse_100101.csv")
	assert.NoError(t, err)
	assert.Len(t, data, 2, "should parse 2 rows")
	assert.Equal(t, "new_order_001", data["new_order_001"]["StepId"])
	assert.Equal(t, "new_order_002", data["new_order_002"]["StepId"])
}
