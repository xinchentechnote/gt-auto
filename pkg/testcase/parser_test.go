package testcase

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCSVCaseParserParse(t *testing.T) {
	filePath := filepath.Join("testdata", "szse_test_case.csv")

	parser := &CSVCaseParser{FilePath: filePath}
	cases, err := parser.Parse()

	assert.NoError(t, err)
	assert.Len(t, cases, 1, "should parse 1 test case")

	tc := cases[0]
	assert.Equal(t, "szse_001", tc.CaseID)
	assert.Equal(t, "order", tc.CaseTitle)
	assert.Len(t, tc.Steps, 4, "should have 4 steps")

	assert.Equal(t, "new_order_001", tc.Steps[0].StepID)
	assert.Equal(t, "szse_bin_oms_1", tc.Steps[0].TestTool)
	assert.Equal(t, "100101", tc.Steps[0].MsgType)
	assert.Equal(t, "new_order_001", tc.Steps[0].TestDatas["StepId"])
	assert.Equal(t, "c0001", tc.Steps[0].TestDatas["ClOrdID"])

	assert.Equal(t, "new_order_002", tc.Steps[1].StepID)
	assert.Equal(t, "szse_bin_tgw_1", tc.Steps[1].TestTool)
	assert.Equal(t, "100101", tc.Steps[1].MsgType)

	assert.Equal(t, "new_order_003", tc.Steps[2].StepID)
	assert.Equal(t, "szse_bin_tgw_1", tc.Steps[2].TestTool)
	assert.Equal(t, "200102", tc.Steps[2].MsgType)
}

func TestLoadCSVToMap(t *testing.T) {
	data, err := LoadCSVToMap("testdata/szse_100101.csv")
	assert.NoError(t, err)
	assert.Len(t, data, 2, "should parse 2 rows")
	assert.Equal(t, "new_order_001", data["new_order_001"]["StepId"])
	assert.Equal(t, "new_order_002", data["new_order_002"]["StepId"])
}

// TestLoadRiskConfirmData locks in the risk_200102.csv header fix: the third
// column was a duplicated UniqueOrderID which silently overwrote the first.
func TestLoadRiskConfirmData(t *testing.T) {
	data, err := LoadCSVToMap("testdata/risk_200102.csv")
	assert.NoError(t, err)
	assert.Equal(t, "ORDERID00001", data["new_order_003"]["UniqueOrderID"])
	assert.Equal(t, "CLORD0001", data["new_order_003"]["UniqueOrigOrderID"])
	assert.Equal(t, "ORIGCLORD1", data["new_order_003"]["ClOrdID"])
}

// TestLoadCSVToMapMissingStepIdColumn verifies a data file without a StepId
// column produces an error instead of a panic.
func TestLoadCSVToMapMissingStepIdColumn(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "data.csv")
	err := os.WriteFile(f, []byte("Foo,Bar\n1,2\n"), 0o644)
	assert.NoError(t, err)
	_, err = LoadCSVToMap(f)
	assert.ErrorContains(t, err, "StepId")
}

// TestParseRejectsShortRow verifies a CSV row with too few columns produces
// an error instead of an index-out-of-range panic.
func TestParseRejectsShortRow(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "case.csv")
	content := "case_id,case_title,step_id,sleep_ms,step_desc,action_type,verify_required,test_tool,msg_type,test_data\n" +
		"c1,title,step_001,1,desc,Send,N,tool,100101,data\n" +
		",,short_row,1,broken\n"
	err := os.WriteFile(f, []byte(content), 0o644)
	assert.NoError(t, err)
	parser := &CSVCaseParser{FilePath: f}
	_, err = parser.Parse()
	assert.ErrorContains(t, err, "columns")
}

// TestFindTestDataSheetIsolation verifies the same StepId in two different
// sheets resolves to each sheet's own record instead of a cached collision.
func TestFindTestDataSheetIsolation(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}
	write("case.csv", "case_id,case_title,step_id,sleep_ms,step_desc,action_type,verify_required,test_tool,msg_type,test_data\n")
	write("sheetA.csv", "StepId,ClOrdID\nshared,from_a\n")
	write("sheetB.csv", "StepId,ClOrdID\nshared,from_b\n")

	parser := &CSVCaseParser{FilePath: filepath.Join(dir, "case.csv")}
	a, err := parser.findTestData("sheetA", "shared")
	assert.NoError(t, err)
	b, err := parser.findTestData("sheetB", "shared")
	assert.NoError(t, err)
	assert.Equal(t, "from_a", a["ClOrdID"])
	assert.Equal(t, "from_b", b["ClOrdID"])
}

func TestFindTestDataStepNotFound(t *testing.T) {
	parser := &CSVCaseParser{FilePath: filepath.Join("testdata", "risk_test_case.csv")}
	_, err := parser.findTestData("risk_100101", "no_such_step")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no_such_step")
}

// TestParseMissingTestDataStep verifies a step whose test data cannot be
// resolved is kept with a SkipReason (the executor turns it into a failure)
// instead of being silently dropped.
func TestParseMissingTestDataStep(t *testing.T) {
	dir := t.TempDir()
	caseFile := filepath.Join(dir, "case.csv")
	dataFile := filepath.Join(dir, "data.csv")
	err := os.WriteFile(dataFile, []byte("StepId,ClOrdID\nstep_001,c0001\n"), 0o644)
	assert.NoError(t, err)
	csvContent := "case_id,case_title,step_id,sleep_ms,step_desc,action_type,verify_required,test_tool,msg_type,test_data\n" +
		"c1,title,step_001,1,desc,Send,N,tool,100101,data\n" +
		",,step_missing,1,desc,Send,N,tool,100101,data\n"
	err = os.WriteFile(caseFile, []byte(csvContent), 0o644)
	assert.NoError(t, err)

	parser := &CSVCaseParser{FilePath: caseFile}
	cases, err := parser.Parse()
	assert.NoError(t, err)
	assert.Len(t, cases, 1)
	assert.Len(t, cases[0].Steps, 2, "unresolvable step must not be dropped")
	resolved, missing := cases[0].Steps[0], cases[0].Steps[1]
	assert.Equal(t, "step_001", resolved.StepID)
	assert.Empty(t, resolved.SkipReason)
	assert.NotNil(t, resolved.TestDatas)
	assert.Equal(t, "step_missing", missing.StepID)
	assert.Contains(t, missing.SkipReason, "step_missing not found in test data file")
	assert.Nil(t, missing.TestDatas)
}
