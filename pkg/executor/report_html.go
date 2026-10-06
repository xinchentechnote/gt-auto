package executor

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

// htmlReportTemplate renders the report as a self-contained HTML page (inline
// CSS, no external assets) so it can be archived or attached anywhere.
const htmlReportTemplate = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<title>GT-Auto 测试报告</title>
<style>
  body { font-family: -apple-system, "Segoe UI", "PingFang SC", sans-serif; margin: 24px; color: #1f2328; }
  h1 { font-size: 20px; margin-bottom: 4px; }
  h2 { font-size: 15px; margin: 26px 0 8px; }
  .meta { color: #57606a; font-size: 13px; margin: 0 0 16px; }
  .cards { display: flex; gap: 14px; flex-wrap: wrap; }
  .card { border: 1px solid #d0d7de; border-radius: 8px; padding: 10px 18px; min-width: 100px; text-align: center; }
  .card .num { font-size: 24px; font-weight: 600; }
  .card .label { font-size: 12px; color: #57606a; }
  .card.failed .num { color: #cf222e; }
  .card.passed .num { color: #1a7f37; }
  table { border-collapse: collapse; width: 100%; margin: 8px 0 4px; font-size: 13px; }
  th, td { border: 1px solid #d0d7de; padding: 6px 10px; text-align: left; vertical-align: top; }
  th { background: #f6f8fa; white-space: nowrap; }
  tr.step-failed td { background: #fff8f6; }
  td.err { color: #cf222e; }
  table.diffs { margin: 4px 0; }
  table.diffs th { background: #fbfbfc; }
</style>
</head>
<body>
<h1>GT-Auto 测试报告</h1>
<p class="meta">开始 {{.StartedAt.Format "2006-01-02 15:04:05"}} · 结束 {{.FinishedAt.Format "2006-01-02 15:04:05"}} · 耗时 {{.DurationMs}} ms</p>
<div class="cards">
  <div class="card"><div class="num">{{.Summary.TotalCases}}</div><div class="label">用例</div></div>
  <div class="card"><div class="num">{{.Summary.TotalSteps}}</div><div class="label">步骤结果</div></div>
  <div class="card passed"><div class="num">{{.Summary.PassedSteps}}</div><div class="label">通过</div></div>
  <div class="card failed"><div class="num">{{.Summary.FailedSteps}}</div><div class="label">失败</div></div>
</div>
{{range .Cases}}
<h2>{{if .Passed}}✅{{else}}❌{{end}} {{.CaseID}}{{if .CaseTitle}} — {{.CaseTitle}}{{end}}</h2>
{{if .Steps}}
<table>
  <tr><th>#</th><th>步骤</th><th>结果</th><th>错误 / 差异</th></tr>
  {{range .Steps}}
  <tr{{if not .Passed}} class="step-failed"{{end}}>
    <td>{{.Index}}</td>
    <td>{{.StepID}}</td>
    <td>{{if .Passed}}✅{{else}}❌{{end}}</td>
    <td>{{if .Error}}<div class="err">{{.Error}}</div>{{end}}
      {{if .Diffs}}
      <table class="diffs">
        <tr><th>路径</th><th>期望</th><th>实际</th></tr>
        {{range .Diffs}}
        <tr><td>{{.Path}}</td><td><pre>{{toStr .Expect}}</pre></td><td><pre>{{toStr .Actual}}</pre></td></tr>
        {{end}}
      </table>
      {{end}}
    </td>
  </tr>
  {{end}}
</table>
{{else}}<p class="meta">本用例无步骤结果（所有 Send 步骤执行成功且无校验步骤）</p>{{end}}
{{end}}
</body>
</html>
`

// WriteHTMLReport renders the report as HTML and writes it to path, creating
// parent directories as needed.
func WriteHTMLReport(report Report, path string) error {
	tmpl, err := template.New("report").Funcs(template.FuncMap{
		"toStr": diffValueToString,
	}).Parse(htmlReportTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse report template: %w", err)
	}
	var buf strings.Builder
	if err := tmpl.Execute(&buf, report); err != nil {
		return fmt.Errorf("failed to render report: %w", err)
	}
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create report directory: %w", err)
		}
	}
	if err := os.WriteFile(path, []byte(buf.String()), 0o644); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}
	return nil
}

// diffValueToString renders a diff value for the HTML report: strings stay
// as-is, structured values become indented JSON so field differences are
// easy to eyeball.
func diffValueToString(v any) string {
	if v == nil {
		return "<nil>"
	}
	if s, ok := v.(string); ok {
		return s
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(data)
}
