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
	app := &cli.App{
		Name:  "gt-auto",
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
			}, &cli.StringSliceFlag{
				Name:  "report",
				Usage: "Report file(s) to write; format follows the extension (.json or .html); repeatable, empty disables",
				Value: cli.NewStringSlice("gt-auto-report.json"),
			},
		},
		Action: func(c *cli.Context) error {
			casePath := c.String("casePath")
			log.Info("Running test from: \n", casePath)
			cases, err := testcase.LoadTestCases(casePath)
			if err != nil {
				return fmt.Errorf("failed to load test cases: %w", err)
			}
			configPath := c.String("config")
			log.Info("Using config from: \n", configPath)
			gwAutoConfig, err := config.ParseConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to parse config: %w", err)
			}
			gwAutoConfig.InitConfigMap()

			caseExecutor := executor.NewCaseExecutor(*gwAutoConfig, cases)
			summary := caseExecutor.Execute()
			log.Infof("Run summary: %d case(s), %d step(s): %d passed, %d failed",
				summary.TotalCases, summary.TotalSteps, summary.PassedSteps, summary.FailedSteps)
			runReport := caseExecutor.BuildReport()
			for _, reportPath := range c.StringSlice("report") {
				if reportPath == "" {
					continue
				}
				if err := executor.WriteReport(runReport, reportPath); err != nil {
					return fmt.Errorf("failed to write report: %w", err)
				}
				log.Infof("Report written to %s", reportPath)
			}
			if summary.FailedSteps > 0 {
				return cli.Exit(fmt.Sprintf("%d step(s) failed", summary.FailedSteps), 1)
			}
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
