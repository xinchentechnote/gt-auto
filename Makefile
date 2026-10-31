# gt-auto 构建与测试编排。
#
# 默认目标为纯 Go（无需 cgo 与外部仓库）；fbsdk 相关目标依赖同级检出的
# fbsdk（Rust cdylib）与 fcounter（被测系统），路径可用变量覆盖。

FBSDK_DIR    ?= $(CURDIR)/../fbsdk
FCOUNTER_DIR ?= $(CURDIR)/../fcounter

BIN         := bin/gt-auto
BIN_FBSDK   := bin/gt-auto-fcounter
GO_TAGS_FBS := -tags fbsdk

.PHONY: help build build-fbsdk install run test test-fbsdk vet vet-fbsdk e2e-fcounter clean

help: ## 列出全部目标
	@grep -E '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

build: ## 构建纯 Go 版二进制（bin/gt-auto；不含 fbsdk 模拟器）
	go build -o $(BIN) ./cmd/gt-auto/

build-fbsdk: ## 构建内嵌 FB-SDK 的二进制（自动先构建 Rust cdylib）
	cd $(FBSDK_DIR) && cargo build -p fbsdk-ffi --release
	CGO_ENABLED=1 CGO_LDFLAGS="-L$(abspath $(FBSDK_DIR)/target/release)" go build $(GO_TAGS_FBS) -o $(BIN_FBSDK) ./cmd/gt-auto/

install: build ## 构建并安装到 ~/go/bin
	cp $(BIN) ~/go/bin/

run: build ## 自回环示例：risk / sse / szse 三协议（gt-auto 双侧模拟器互打）
	./$(BIN) --casePath pkg/testcase/testdata/risk_test_case.csv --config pkg/config/testdata/gw-auto-risk.toml
	./$(BIN) --casePath pkg/testcase/testdata/sse_test_case.csv --config pkg/config/testdata/gw-auto-sse.toml
	./$(BIN) --casePath pkg/testcase/testdata/szse_test_case.csv --config pkg/config/testdata/gw-auto-szse.toml

test: ## 全量测试（纯 Go；-race）
	go test -race -v ./...

test-fbsdk: ## FB-SDK 模拟器测试（自动构建 cdylib；-tags fbsdk -race）
	cd $(FBSDK_DIR) && cargo build -p fbsdk-ffi --release
	CGO_ENABLED=1 CGO_LDFLAGS="-L$(abspath $(FBSDK_DIR)/target/release)" go test $(GO_TAGS_FBS) -race ./...

vet: ## 静态检查（纯 Go）
	go vet ./...

vet-fbsdk: ## 静态检查（-tags fbsdk）
	go vet $(GO_TAGS_FBS) ./...

e2e-fcounter: ## fcounter 全链路验收（fb-sdk ↔ fcounter ↔ SSE；需同级 fcounter/fbsdk 检出）
	@test -d $(FCOUNTER_DIR)/.git || { echo "缺少同级 fcounter 检出：$(FCOUNTER_DIR)"; exit 1; }
	@test -d $(FBSDK_DIR)/.git || { echo "缺少同级 fbsdk 检出：$(FBSDK_DIR)"; exit 1; }
	cd $(FCOUNTER_DIR) && ./scripts/e2e_gt_auto.sh

clean: ## 清理构建产物
	rm -rf bin/
