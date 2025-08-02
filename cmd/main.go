package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/xinchentechnote/gt-auto/pkg/config"
	"github.com/xinchentechnote/gt-auto/pkg/executor"
	"github.com/xinchentechnote/gt-auto/pkg/testcase"
)

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := filepath.Base(f.File)
			funcName := f.Function
			parts := strings.Split(funcName, "/")
			shortFunc := parts[len(parts)-1]
			return fmt.Sprintf("%s()", shortFunc),
				fmt.Sprintf("%s:%d", filename, f.Line)
		},
	})
	log.SetReportCaller(true)
	log.SetLevel(log.InfoLevel)
	log.SetReportCaller(true)
	app := &cli.App{
		Name:  "gw-auto",
		Usage: "CLI tool for gateway automation testing",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "casePath",
				Usage:    "Path to the test case file path",
				Required: true,
			}, &cli.StringFlag{
				Name:     "config",
				Usage:    "Path to the configuration file",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			// 1.Parse test cases from the provided file
			casePath := c.String("casePath")
			log.Info("Running test from: \n", casePath)
			cases, err := testcase.LoadTestCases(casePath)
			if err != nil {
				panic(err)
			}
			configPath := c.String("config")
			log.Info("Using config from: \n", configPath)
			// 2. Create a simulators based on the configuration
			gwAutoConfig, err := config.ParseConfig(configPath)
			if err != nil {
				panic(err)
			}
			gwAutoConfig.InitConfigMap()
			// 3. Execute the test cases
			executor := executor.NewCaseExecutor(*gwAutoConfig, cases)
			// 4. Collect the results,validate and generate a report
			executor.Execute()
			// 5. Save the report to a file
			// 6. Print the report to the console
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err)
	}

}
