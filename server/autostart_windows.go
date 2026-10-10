//go:build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/windows/registry"
)

const (
	autostartRunKey    = `Software\Microsoft\Windows\CurrentVersion\Run`
	autostartValueName = "NextPaste"
)

// setAutostart 开启/关闭开机自启（Windows：HKCU 注册表 Run 键，纯 Go 无 CGO 依赖）
// 开启时写入：`"<exe 绝对路径>" -minimized -autostart=<server|client>`
func setAutostart(enabled bool, mode AutoStartMode) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, autostartRunKey, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("打开注册表 Run 键失败: %w", err)
	}
	defer k.Close()

	if !enabled {
		if err := k.DeleteValue(autostartValueName); err != nil && !errors.Is(err, registry.ErrNotExist) {
			return fmt.Errorf("移除开机自启注册项失败: %w", err)
		}
		return nil
	}

	execPath, err := getExecPath()
	if err != nil {
		return fmt.Errorf("获取程序路径失败: %w", err)
	}
	if execPath == "" {
		return errors.New("程序路径为空")
	}
	if isDevBuild(execPath) {
		return errors.New("当前运行的是开发模式构建（wails dev），注册开机自启会指向临时文件。请先 wails build 出正式版再开启")
	}

	value := fmt.Sprintf(`"%s" -minimized -autostart=%s`, execPath, mode)
	if err := k.SetStringValue(autostartValueName, value); err != nil {
		return fmt.Errorf("写入开机自启注册项失败: %w", err)
	}
	return nil
}

// isAutostartEnabled 查询注册表中的自启状态，返回 (是否启用, 自启模式, 错误)
func isAutostartEnabled() (bool, AutoStartMode, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, autostartRunKey, registry.QUERY_VALUE)
	if err != nil {
		return false, AutoStartModeServer, fmt.Errorf("打开注册表 Run 键失败: %w", err)
	}
	defer k.Close()

	v, _, err := k.GetStringValue(autostartValueName)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return false, AutoStartModeServer, nil
		}
		return false, AutoStartModeServer, err
	}

	mode := AutoStartModeServer
	if strings.Contains(v, "-autostart=client") {
		mode = AutoStartModeClient
	}
	return true, mode, nil
}

// isDevBuild 判断是否运行在开发模式（wails dev）构建产物中：
// wails dev 的二进制位于系统临时目录，指向它的自启注册项无意义
func isDevBuild(execPath string) bool {
	lower := strings.ToLower(execPath)
	tmp := strings.ToLower(os.TempDir())
	return strings.HasPrefix(lower, tmp) || strings.Contains(lower, `\temp\wails\`)
}
