package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/xinchentechnote/gt-auto/pkg/testcase"
)

func main() {
	app := &cli.App{
		Name:  "gw-auto",
		Usage: "CLI tool for gateway automation testing",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "casePath",
				Usage:    "Path to the test case file path",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			// 1.Parse test cases from the provided file
			casePath := c.String("casePath")
			fmt.Printf("Running test from: %s\n", casePath)
			cases, err := testcase.LoadTestCases(casePath)
			if err != nil {
				panic(err)
			}

			for _, c := range cases {
				fmt.Printf("CaseNo: %s | Title: %s\n", c.CaseNo, c.CaseTitle)
				for _, s := range c.Steps {
					fmt.Printf("  Step: %-20s | Action: %-6s | Tool: %-25s | Data: %s\n",
						s.StepID, s.ActionType, s.TestTool, s.TestDataSheet)
				}
			}
			// 2. Create a simulators based on the configuration
			// 3. Execute the test cases
			// 4. Collect the results,validate and generate a report
			// 5. Save the report to a file
			// 6. Print the report to the console
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
