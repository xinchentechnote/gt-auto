## Overview
GT-Auto is a gateway auto test tool designed for financial trading systems. It provides a framework for automating the testing of message exchanges between an Order Management System (OMS) and a Trading Gateway (TGW).

## Documentation
- [架构文档 (Architecture)](docs/architecture.md) – 分层设计、核心接口、执行流程、并发模型、协议帧格式
- [开发设计文档 (Design)](docs/design.md) – 设计决策（ADR）、用例/配置格式规范、扩展指南、测试策略、路线图

## Quick Start
```bash
make build
./bin/gt-auto --casePath pkg/testcase/testdata/risk_test_case.csv --config pkg/config/testdata/gw-auto-risk.toml
# exit code 0 = all steps passed; 1 = at least one step failed
```
- Here is the system architecture diagram for the production environment:
```mermaid
graph LR
    OMS["OMS<br>(Order Management System)"]
    GW1["GW<br>(Trading Gateway)"]
    GW2["GW<br>(Risk Control Gateway)"]
    TGW["TGW<br>(Exchange)"]

    OMS --> GW1
    OMS --> GW2
    GW1 --> TGW
    GW2 --> TGW
```
- System architecture diagram for the test environment:
```mermaid
graph LR
    OMS[OmsSimulator]
    GW[Target Gateway]
    TGW[TgwSimulator]

    OMS --> GW
    GW --> TGW
```
- System architecture diagram for the test environment with a focus on the input and output components:
```mermaid
graph LR
    Input[InputSimulator]
    Component[Target Component]
    Output[OutputSimulator]

    Input --> Component
    Component --> Output
```
##  Supported Protocol Types
- [x] **RiskBin** - Risk Control Binary Protocol, used for high-speed risk control data exchange.
- [x] **SzseBin** – Shenzhen Stock Exchange Binary Protocol, used for high-speed market data or trading access.
- [ ] **SzseStep** – Shenzhen Stock Exchange STEP Protocol, supports richer session and order interaction.
- [x] **SseBin** – Shanghai Stock Exchange Binary Protocol, similar to SzseBin, used for trading and data exchange.
- [ ] **SseStep** – Shanghai Stock Exchange STEP Protocol, provides comprehensive support for order and execution data.
- [ ] **FIX** (Financial Information eXchange) – Widely used international standard for communication between traders, brokers, and exchanges.
- [ ] **IMIX** (Inter-bank Market Information eXchange) – Protocol for communication between financial institutions in interbank markets.
- [ ] **Protobuf** – Google Protocol Buffers used for efficient service-to-service communication.
- [ ] **Custom** – User-defined protocol (binary or text-based), tailored for specific business needs.

[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/xinchentechnote/gt-auto)