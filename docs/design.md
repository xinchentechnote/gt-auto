# GT-Auto 开发设计文档

> 更新日期：2026-10-06。架构与组件关系见[架构文档](architecture.md)；本文记录设计决策、数据格式规范、扩展指南与测试策略。

## 1. 设计目标

1. **结果可信**：所有失败（比对差异、模拟器故障、用例数据错误）都必须体现在报告中，杜绝"假通过"。
2. **永不卡死**：任何外部异常（网关静默、连接断开、坏数据）都转化为明确错误，不允许 panic 或无限阻塞。
3. **协议可扩展**：新增一个二进制协议只需实现 codec + framer 并注册，执行编排与模拟器零改动。
4. **CI 友好**：非零退出码表达失败；测试全部可在 `-race` 下运行。

## 2. 关键设计决策（ADR）

### D1. Framer 与 Codec 分离

- **背景**：TCP 是字节流，"切帧"（读定长头 + 按长度读体）与"解码"（字节 → 结构体）是两个独立关注点；不同交易所帧头不同，但解码生态（`BinaryCodec`）相同。
- **决策**：`Framer` 只负责 `ReadFrame(conn) ([]byte, error)`；`MessageCodec` 只负责编解码。模拟器持有两者，串联 `ReadFrame → Decode`。
- **后果**：新增协议时两者各自实现、独立测试；帧格式问题与字段映射问题可分别定位。

### D2. 泛型模拟器 `Simulator[T fin_codec.BinaryCodec]`

- **决策**：模拟器对报文体类型泛型化（`CreateSimulator[T]`），执行器统一实例化为 `tcp.Simulator[codec.BinaryCodec]`。
- **后果**：接口返回 `T` 而非 `interface{}`；队列取出的类型断言集中在一处（`dequeueWithContext`）。

### D3. 接收队列：带缓冲 channel（替换 goconcurrentqueue）

- **背景**：早期用 `goconcurrentqueue.FIFO`，超时取数依赖其 context 版 API，且存在遗留 watcher 的实现细节。
- **决策**：`queue chan receivedMessage`（容量 1024，构造时创建、之后不可变），`Receive` 用 `select { case <-queue; case <-timer.C }` 实现超时。
- **后果**：删除一个第三方依赖；超时语义一目了然；字段不可变后无并发写。
- **元素设计**：`receivedMessage{MsgType, Body}` 而非裸 body —— 线上消息类型一路带到校验层（见 D6）。

### D4. 就绪轮询替代 `time.Sleep`

- **背景**：早期用固定 sleep（每模拟器 1s、用例开始前 5s、每步骤 1s）做启动同步：慢且不可靠。
- **决策**：`Simulator.Ready()`（OMS=已拨号 / TGW=listener 已绑定）+ executor `waitReady`（10ms 轮询、5s 超时）。`auto_start` 模拟器在 init 阶段等待就绪，惰性模拟器在首次使用时启动并等待。
- **后果**：启动时间从"至少 N 秒"变为"实际就绪时间"；未就绪（如拨号被拒）直接让步骤失败并记录原因。步骤间等待完全由用例的 `sleep_ms` 决定。

### D5. Receive 对 MsgType 快速失败（而非过滤等待）

- **背景**：早期队列丢弃线上 MsgType，网关回错报文时只能得到一个难懂的整体类型不匹配 diff。
- **决策**：队列元素携带 MsgType；Receive 返回 `(body, msgType, err)`；执行器比对步骤期望的 MsgType，不符立即记为步骤失败（`received MsgType X, expected Y`）。
- **取舍**：不实现"过滤等待 + 暂存不匹配报文"。当前被测流程是请求/响应式，回错类型本身就是要暴露的缺陷；过滤会引入报文丢弃与乱序语义。若未来需要支持心跳/多路消息，在此处扩展。

### D6. 错误处理约定

| 规则 | 落点 |
|---|---|
| 库代码不 `log.Fatal`/`os.Exit`，只返回 error | `config.ParseConfig`（曾用 `log.Fatalf` 导致 defer 失效） |
| 包装错误必须包装**真实原因**，用 `%w` 保持 `errors.Is/As` 链路 | 三个 framer、`receive0`（曾包装 nil 变量丢失原因） |
| 外部输入（CSV、JSON map）不允许裸类型断言，缺失/错型返回明确错误 | `msgTypeFromMap`、`LoadCSVToMap`、用例行长度检查 |
| 连接不可恢复错误（EOF/UnexpectedEOF/ErrClosed）终止接收循环，单条解码失败仅跳过 | `receiveLoop` / `handleClient` |
| 步骤无法执行 ≠ 静默跳过：一律 `AddStepError` 落入报告 | `executeStep` 全部分支；解析器对无法构建的步骤保留并标记 `SkipReason`，由执行器记为失败 |

### D7. 测试数据按 sheet 缓存

- **决策**：`CSVCaseParser.sheetCache[sheetName] → (stepID → record)`，每个数据文件最多读一次；记录查找只在本 sheet 内进行。
- **原因**：早期缓存以 stepID 为全局键，两个 sheet 出现同名 StepId 时先加载者胜出（串数据）；且缓存未命中会反复读同一文件。

### D8. 失败必须可见

- 模拟器创建失败：init 阶段 Error 日志 + 惰性路径让步骤失败。
- 测试数据无法加载：解析阶段 Warn + 跳过该步骤（步骤缺失会在结果中表现为步骤数少于预期）。
- 未知 ActionType / 工具名不在配置：步骤失败，错误信息包含实际值（如 `test tool "xxx" not found in config simulators`）。

## 3. 测试用例格式规范

### 3.1 主用例文件（CSV，UTF-8，首行表头，固定 10 列）

| 列 | 字段 | 说明 |
|---|---|---|
| 1 | `case_id` | 用例 ID；**非空即开启新用例**，后续行留空表示归属当前用例 |
| 2 | `case_title` | 用例标题 |
| 3 | `step_id` | 步骤 ID；同时是测试数据 sheet 中的行键 |
| 4 | `sleep_ms` | 执行本步骤前的等待毫秒数；空/非法/≤0 不等待 |
| 5 | `step_desc` | 步骤描述（仅日志展示） |
| 6 | `action_type` | `Send` 或 `Receive`；其他值记为步骤失败 |
| 7 | `verify_required` | `Y`（大小写不敏感）时 Receive 步骤执行字段级比对 |
| 8 | `test_tool` | 模拟器名，必须与配置中的 `name` 一致 |
| 9 | `msg_type` | 期望的线上消息类型（十进制字符串，如 `100101`、`58`） |
| 10 | `test_data` | 数据 sheet 名（不含扩展名），工具在同目录找 `<sheet><ext>` 文件 |

示例（`pkg/testcase/testdata/risk_test_case.csv`）：

```csv
case_id,case_title,step_id,sleep_ms,step_desc,action_type,verify_required,test_tool,msg_type,test_data
risk_001,order,new_order_001,1,oms send new order,Send,N,risk_bin_oms_1,100101,risk_100101
,,new_order_002,1,tgw receive new order,Receive,Y,risk_bin_tgw_1,100101,risk_100101
,,new_order_003,1,tgw send confirm,Send,N,risk_bin_tgw_1,200102,risk_200102
,,new_order_004,1,oms receive confirm,Receive,Y,risk_bin_oms_1,200102,risk_200102
```

典型四步流：OMS 发单 → TGW 收到（比对）→ TGW 发确认 → OMS 收到（比对）。

### 3.2 数据 sheet（CSV）

- 首行表头 = 报文字段名（须与 proto 结构体的 `json` tag 一致，如 `ClOrdID`、`UniqueOrigOrderID`）；**必须包含 `StepId` 列**作为行键。
- 每行是一条完整的步骤数据；执行器按主用例的 `step_id` 在 sheet 中查找同名行。
- 查找不到 → 该步骤标记 `SkipReason`，执行时记为失败。
- Send 与 Receive 引用同一行：Receive 的期望值即"网关应原样转发/回执该报文"。

**跨文件嵌套**：数据表单元格值以 `@` 开头表示引用另一个数据文件，**由写法直接区分对象与数组**：
- `@parts` → **嵌套对象**：要求 `parts.csv` 恰好 1 行数据；多行报错（提示改用 `@parts[]`）、空行报错，不做猜测；
- `@parts[]` → **对象数组**：取全部行（行序即数组顺序，单行也是单元素数组、无数据行为空数组）。

引用可递归（带环检测，成环报错），被引用文件同样必须含 `StepId` 列（值任意，如 `p1`/`p2`）。例：

```
order.csv:  StepId,ClOrdID,Partition
            s1,c0001,@parts[]       ← Partition 字段引用 parts.csv（数组模式）
parts.csv:  StepId,PlatformID
            p1,101
            p2,202
→ 展开后 TestDatas.Partition = [{PlatformID:101},{PlatformID:202}]，填充 []*PlatformPartition 等列表字段
```

### 3.3 JSON 用例格式（.json，数据内联）

与 CSV 等价的替代格式：`testData` 对象直接内联在每个步骤里，**无需独立的数据 sheet 文件**，适合程序化生成用例。字段名与 CSV 列一一对应（驼峰式）；`msgType`/`sleepMs` 接受字符串或数字；`verifyRequired` 为布尔。示例见 `pkg/testcase/testdata/risk_test_case.json`：

```json
{
  "cases": [
    {
      "caseId": "risk_001",
      "caseTitle": "order",
      "steps": [
        {
          "stepId": "new_order_001",
          "sleepMs": 1,
          "stepDesc": "oms send new order",
          "actionType": "Send",
          "verifyRequired": false,
          "testTool": "risk_bin_oms_1",
          "msgType": "100101",
          "testData": {"ClOrdID": "c00001", "Side": "1"}
        }
      ]
    }
  ]
}
```

### 3.4 Excel 用例格式（.xlsx，单工作簿）

单工作簿承载全部内容，布局与 CSV 完全对应：

- **第一个 sheet 为用例表**：与 CSV 主文件相同的 10 列表头与填法（`case_id` 非空开新用例）；
- **其余 sheet 为数据表**：sheet 名即 `test_data` 引用值，首行为字段表头（必须含 `StepId` 列），行按 `StepId` 查找——与 CSV 的数据 sheet 文件布局一致。

约定与容错：尾随空单元格自动补空；全空行跳过；引用不存在的数据 sheet 时该步骤保留并标记 `SkipReason`（执行时记为失败）；旧版 `.xls` 二进制格式不支持（用 excelize 另存为 `.xlsx`）。示例见 `pkg/testcase/testdata/risk_test_case.xlsx`。

**跨 sheet 嵌套**：与 CSV 的跨文件嵌套同一约定——数据 sheet 单元格值以 `@` 开头引用同工作簿的另一个数据 sheet：`@parts` 对象（恰好 1 行）、`@parts[]` 对象数组（全部行），递归 + 环检测。

## 4. 配置文件规范（TOML）

```toml
# 可选：Receive 步骤等待报文的超时（毫秒），缺省 5000
receive_timeout_ms = 3000

[[simulators]]
name = "szse_bin_tgw_1"        # 必须唯一；与用例的 test_tool 对应
type = "tgw"                    # oms | tgw
communication = "tcp"           # 当前仅 tcp（字段保留）
protocol = "binary-szse"        # binary-risk | binary-szse | binary-sse
listen_address = ":9003"        # tgw 必填
auto_start = true               # true: 随工具启动并等待就绪; false: 首次被用例引用时惰性启动

[[simulators]]
name = "szse_bin_oms_1"
type = "oms"
communication = "tcp"
protocol = "binary-szse"
server_address = "localhost:9003"  # oms 必填：被测网关监听地址
auto_start = false
```

经验配置：tgw 用 `auto_start = true`（先监听，网关才能接入）；oms 用 `auto_start = false`（首次发单时才拨号）。

## 5. 扩展指南

### 5.1 新增一个二进制协议（以 `binary-xyz` 为例）

1. **报文库**：提供 `XyzBinary{MsgType, Body}` 封包结构与 `NewXyzBinaryMessageByMsgType` 工厂（参考 `fin-proto-szse-bin-go`，通常代码生成）。
2. **Codec**：在 `pkg/codec` 新增 `XyzMessageCodec`，实现 `MessageCodec` 五个方法；`MsgType` 解析复用 `msgTypeFromMap`，map→struct 复用 `ConvertMapToStruct`。`Encode(msgType uint32, message)` 负责封包头尾。
3. **Framer**：新增 `XyzBinFramer`，按帧头布局 `io.ReadFull` 读取；**包装 body 读取错误时 wrap 实际的 error 变量**（历史教训）。
4. **注册**：工厂 `GetCodec`/`GetFramer` 各加一个 case，常量加进 `message_codec_factory.go`。
5. **测试**：帧截断错误路径（参考 `framer_test.go` 的表驱动）+ 编解码往返。

执行编排、模拟器、用例格式均无需改动。

### 5.2 新增模拟器类型（如 udp）

`tcp.CreateSimulator` 的 `config.Type` switch 加分支，实现 `Simulator[T]` 接口（重点是 `Ready` 的就绪语义与收发路径）。

### 5.3 新增用例格式

实现 `testcase.CaseParser` 接口，在 `LoadTestCases` 按扩展名挂接；产物为统一的 `[]*TestCase` 模型，下游无感知。行→步骤构建（`parseCaseRows`）与数据表查找（`sheetCache.lookup` / `recordsFromRows`）已抽为共享核心，表格式新格式（如 Google Sheets 导出）可直接复用；`JSONCaseParser` 是内联数据格式的参考实现。

## 6. 测试策略

| 层次 | 位置 | 内容 |
|---|---|---|
| codec 单测 | `pkg/codec/*_test.go` | map→struct 转换、非法 MsgType 表驱动（3 协议 × 3 形态）、帧截断错误路径 |
| 模拟器单测 | `pkg/tcp/receive_loop_test.go` | 断连退出、Receive 超时、Close 幂等/未启动安全、无连接发送报错 |
| 集成测试 | 同上 `TestOmsTgwEndToEnd` | 真实 TCP 双向收发（risk 协议），断言线上 MsgType |
| executor 单测 | `pkg/executor/case_executor_test.go` | stub simulator 驱动的 Receive/Send 失败路径、汇总计数、就绪等待、auto_start 语义 |
| 解析单测 | `pkg/testcase/parser_test.go` | 用例解析、数据隔离、缺 StepId 列、短行、查无数据 |

约定：

- 提交前 `go test ./... -race`（本项目多次靠 race detector 抓出真实竞争）。
- 修复类提交尽量携带"旧代码必失败"的回归测试（framer 错误包装即先反向验证过）。
- 不依赖真实外网/固定端口：监听一律 `127.0.0.1:0` 后读取实际地址。

## 7. 已知限制与路线图

| 项 | 现状 | 方向 |
|---|---|---|
| Receive 只做 MsgType 快速失败 | 不匹配即失败 | 支持"按类型过滤等待 + 暂存"，适配心跳/多路消息 |
| 报告格式 | JSON 与 HTML 均已落地（`--report` 按扩展名分发、可重复传多份） | — |
| OMS 无重连 | 断连后步骤失败可见 | 增加可配置重连 |
| TGW 多客户端 | Send 发往最近接受的连接 | 按客户端路由 |
| communication 字段 | 仅 tcp 实现 | udp / http |
| 协议覆盖 | risk / szse-bin / sse-bin | STEP(SZSE/SSE)、FIX、IMIX、Protobuf（见 readme 路线图） |
| 嵌套/列表字段 | 已支持：`@` 跨 sheet/文件引用（对象/数组）、JSON 内联嵌套，`convertValue` 逐元素宽松转换（含指针元素） | 接口类型字段（如 `ApplExtend`）仍走 ApplID 注册表机制 |
