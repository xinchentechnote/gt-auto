# testdata —— 测试资产清单与归属约定

本目录的文件分两类，**归属与同步规则不同**：

## 1. 工具自测样例（本仓维护）

gt-auto 自身能力的覆盖样例，随 `go test ./...` 与 `make run` 消费：

| 文件 | 覆盖点 |
|---|---|
| `risk_test_case.csv` + `risk_*.csv` + `gw-auto-risk.toml` | 风控协议双模拟器回环（OMS 发单 → TGW 收单 → TGW 回确认 → OMS 收确认） |
| `sse_test_case.csv` + `sse_*.csv` + `gw-auto-sse.toml` | SSE 协议同构回环 |
| `szse_test_case.csv` + `szse_*.csv` + `gw-auto-szse.toml` | 深市协议同构回环 |
| `fcounter_test_case.csv` + `fcounter_*.csv` + `gw-auto-fcounter.toml` | **fbsdk 模拟器**（cgo）+ SSE 交易所侧的双模拟器编排；覆盖多用例解析、fbsdk 类型/协议/会话字段（`client_id`）配置解析（`TestLoadFcounterCase` / `TestFcounterConfigFixture` / `TestFbsdkJSONToStructDispatch`） |
| `fcounter_szse_test_case.csv` + `fcounter_szse_*.csv` + `gw-auto-fcounter-szse.toml` | 深市（SZSE 二进制）同构场景；ApplExtend 嵌套扩展不做字段级比对（柜台侧携带缺省 tag，flat CSV 无法表达——见 fcounter_002 order_002 注） |

## 2. 被测系统验收用例（权威副本在 fcounter 仓）

`fcounter_*` 套件同时是 **fcounter 的发布验收标准**（受理确认先行、回报序号分配、
OrdType/ClOrdID 映射、0x0020 撤单拒绝码等 counter-design 契约）。**权威副本维护在
[fcounter 仓](https://github.com/xinchentechnote/fcounter) 的 `e2e/gtauto/`**，
由其 `scripts/e2e_gt_auto.sh` 对真柜台执行。

### 同步规则

* counter 协议/字段映射/回报语义变更 → 先改 fcounter 仓的权威副本（随 counter PR），
  再同步本目录副本（本仓 PR），两侧用例内容保持一致。
* 本目录副本仅在 `-tags fbsdk` 且对真柜台运行时执行完整链路；默认 `go test` 只覆盖
  解析与构造（无网络、无 cgo）。

## 运行方式

```sh
go test ./...                                            # 纯 Go：解析/配置/构造覆盖
make test-fbsdk                                          # 含 fbsdk 模拟器（需 Rust cdylib）
make e2e-fcounter                                        # 全链路验收（对真 counter-app 执行）
```
