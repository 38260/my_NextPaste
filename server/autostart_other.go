//go:build !windows

package main

import (
	"errors"
	"fmt"

	"github.com/emersion/go-autostart"
)

// setAutostart 开启/关闭开机自启（macOS/Linux：沿用 go-autostart，无 CGO 依赖）
func setAutostart(enabled bool, mode AutoStartMode) error {
	execPath, err := getExecPath()
	if err != nil {
		return fmt.Errorf("获取程序路径失败: %w", err)
	}
	if execPath == "" {
		return errors.New("程序路径为空")
	}

	app := &autostart.App{
		Name:        "NextPaste",
		DisplayName: "NextPaste Sync Tool",
		Exec:        []string{execPath, "-minimized", "-autostart=" + string(mode)},
	}

	if enabled {
		return app.Enable()
	}
	return app.Disable()
}

// isAutostartEnabled 非 Windows 平台暂不支持状态查询（go-autostart 无查询 API），
// 前端收到错误后保持本地开关状态即可
func isAutostartEnabled() (bool, AutoStartMode, error) {
	return false, AutoStartModeServer, errors.New("当前平台暂不支持查询开机自启状态")
}
