package tcp

import (
	"fmt"

	"github.com/xinchentechnote/gt-auto/pkg/config"
)

// FbsdkSimulatorHook —— fbsdk 客户端模拟器构造钩子（依赖 cgo 的 fbsdk-ffi
// cdylib），`-tags fbsdk` 构建时由 fbsdk 适配包注入；默认构建保持纯 Go。
var FbsdkSimulatorHook func(config config.SimulatorConfig) (any, error)

// newFbsdkSimulatorMissing —— 统一的未启用错误。
func newFbsdkSimulatorMissing() error {
	return fmt.Errorf("fbsdk 模拟器需要 cgo 构建：go build -tags fbsdk（并先构建 fbsdk-ffi cdylib）")
}
