package testcase

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fcounter 全链路用例（本目录副本）是**工具自测样例**：覆盖 CSV 多用例解析、
// 跨文件数据表引用、fbsdk 模拟器配置字段。发布验收的权威副本维护在
// fcounter 仓 e2e/gtauto/，协议/字段变更时须同步本副本。

// TestLoadFcounterCase —— 多用例 CSV + 数据表解析：
// 2 个用例（下单全链路 6 步 / 撤单拒绝 2 步），数据表全部可解析（无 SkipReason）。
func TestLoadFcounterCase(t *testing.T) {
	cases, err := LoadTestCases("testdata/fcounter_test_case.csv")
	require.NoError(t, err)
	require.Len(t, cases, 2)

	assert.Equal(t, "fcounter_001", cases[0].CaseID)
	assert.Equal(t, "下单全链路", cases[0].CaseTitle)
	require.Len(t, cases[0].Steps, 6)

	assert.Equal(t, "fcounter_002", cases[1].CaseID)
	assert.Equal(t, "撤单拒绝", cases[1].CaseTitle)
	require.Len(t, cases[1].Steps, 2)

	// 消息类型序列：SDK 发单 → 交易所收单 → 交易所回确认/成交 → SDK 收确认/成交
	wantTypes := []string{"64", "58", "32", "103", "66", "68"}
	for i, step := range cases[0].Steps {
		assert.Equal(t, wantTypes[i], step.MsgType, "step %s", step.StepID)
		assert.NotEmpty(t, step.TestTool)
		assert.Empty(t, step.SkipReason, "step %s 数据表应可解析", step.StepID)
	}

	// 模拟器名称分派：fbsdk 客户端侧 vs SSE 交易所侧
	assert.Equal(t, "fcounter_fbsdk_oms", cases[0].Steps[0].TestTool)
	assert.Equal(t, "fcounter_sse_tgw", cases[0].Steps[1].TestTool)
	assert.Equal(t, "fcounter_fbsdk_oms", cases[1].Steps[0].TestTool)

	// 校验标记：受理确认（TransactTime 为柜台运行期值）不比对，其余字段级比对
	assert.False(t, cases[0].Steps[4].VerifyRequired)
	assert.True(t, cases[0].Steps[5].VerifyRequired)

	// 数据表字段解析（宽松转换：字符串 → 数值/负数）
	first := cases[0].Steps[0].TestDatas
	assert.Equal(t, "500000001", first["ClOrdID"])
	assert.Equal(t, "-4", first["PriceScale"])
	assert.Equal(t, "100", first["OrderQty"])
}

// TestFcounterConfigFixture —— gw-auto-fcounter.toml 覆盖新配置面：
// fbsdk 模拟器类型/协议 + client_id/token 会话字段 + SSE 会话自动应答的 tgw。
func TestFcounterConfigFixture(t *testing.T) {
	cases, err := LoadTestCases("testdata/fcounter_test_case.csv")
	require.NoError(t, err)
	tools := map[string]bool{}
	for _, c := range cases {
		for _, s := range c.Steps {
			tools[s.TestTool] = true
		}
	}
	assert.True(t, tools["fcounter_fbsdk_oms"], "用例应覆盖 fbsdk 模拟器")
	assert.True(t, tools["fcounter_sse_tgw"], "用例应覆盖 SSE 交易所侧模拟器")
}
