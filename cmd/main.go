package main

import (
	"fmt"

	"github.com/xinchentechnote/gt-auto/pkg/parser"
)

func main() {
	parserImpl := &parser.CSVCaseParser{FilePath: "pkg/parser/testdata/test_case.csv"}
	cases, err := parserImpl.Parse()
	if err != nil {
		panic(err)
	}

	for _, c := range cases {
		fmt.Printf("CaseNo: %s | Title: %s\n", c.CaseNo, c.CaseTitle)
		for _, s := range c.Steps {
			fmt.Printf("  Step: %-20s | Action: %-6s | Tool: %-25s | Data: %s\n",
				s.StepId, s.ActionType, s.TestTool, s.TestDataSheet)
		}
	}
}
