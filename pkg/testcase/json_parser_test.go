package testcase

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONCaseParserParse(t *testing.T) {
	parser := &JSONCaseParser{FilePath: filepath.Join("testdata", "risk_test_case.json")}
	cases, err := parser.Parse()

	assert.NoError(t, err)
	assert.Len(t, cases, 1)

	tc := cases[0]
	assert.Equal(t, "risk_001", tc.CaseID)
	assert.Equal(t, "order", tc.CaseTitle)
	assert.Len(t, tc.Steps, 4)

	// string msgType and numeric msgType both normalize to the model's string
	assert.Equal(t, "100101", tc.Steps[0].MsgType)
	assert.Equal(t, "200102", tc.Steps[3].MsgType)
	assert.Equal(t, "1", tc.Steps[0].SleepMs)
	assert.Equal(t, "risk_bin_oms_1", tc.Steps[0].TestTool)
	assert.False(t, tc.Steps[0].VerifyRequired)
	assert.True(t, tc.Steps[1].VerifyRequired)
	assert.Equal(t, "c00001", tc.Steps[0].TestDatas["ClOrdID"])
}

func TestJSONCaseParserNumbersAndErrors(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) string {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
		return path
	}

	t.Run("malformed json", func(t *testing.T) {
		path := write("bad.json", `{"cases": [`)
		_, err := (&JSONCaseParser{FilePath: path}).Parse()
		assert.ErrorContains(t, err, "failed to parse JSON case file")
	})

	t.Run("wrong sleepMs type", func(t *testing.T) {
		path := write("wrong_sleep.json", `{"cases": [{"caseId": "c1", "steps": [
			{"stepId": "s1", "sleepMs": [1], "actionType": "Send", "testTool": "t", "msgType": "1"}
		]}]}`)
		_, err := (&JSONCaseParser{FilePath: path}).Parse()
		assert.ErrorContains(t, err, "sleepMs")
	})

	t.Run("wrong msgType type", func(t *testing.T) {
		path := write("wrong_msg.json", `{"cases": [{"caseId": "c1", "steps": [
			{"stepId": "s1", "actionType": "Send", "testTool": "t", "msgType": true}
		]}]}`)
		_, err := (&JSONCaseParser{FilePath: path}).Parse()
		assert.ErrorContains(t, err, "msgType")
	})

	t.Run("missing file", func(t *testing.T) {
		_, err := (&JSONCaseParser{FilePath: filepath.Join(dir, "no_such.json")}).Parse()
		assert.ErrorContains(t, err, "failed to read case file")
	})
}

// TestLoadTestCasesJSON verifies the .json dispatch in LoadTestCases and that
// the parsed model is executable-shaped (every step carries test data).
func TestLoadTestCasesJSON(t *testing.T) {
	cases, err := LoadTestCases(filepath.Join("testdata", "risk_test_case.json"))
	assert.NoError(t, err)
	assert.Len(t, cases, 1)
	for _, step := range cases[0].Steps {
		assert.NotNil(t, step.TestDatas)
	}
}

func TestLoadTestCasesUnsupportedExtension(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "case.yaml")
	assert.NoError(t, os.WriteFile(path, []byte("cases: []"), 0o644))
	_, err := LoadTestCases(path)
	assert.ErrorContains(t, err, "unsupported file extension")
}
