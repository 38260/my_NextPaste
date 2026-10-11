//go:build windows

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/sys/windows/registry"
)

const (
	autostartRunKey    = `Software\Microsoft\Windows\CurrentVersion\Run`
	autostartValueName = "ChuanJian"

	// legacyAutostartValueName 项目更名前使用的注册表值名，指向旧 exe 文件名，
	// 更名后必须移除，否则开机可能拉起旧版程序
	legacyAutostartValueName = "NextPaste"
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

	mode := parseAutostartMode(v)
	return true, mode, nil
}

// parseAutostartMode 从自启注册项命令行参数解析工作模式
func parseAutostartMode(value string) AutoStartMode {
	if strings.Contains(value, "-autostart=client") {
		return AutoStartModeClient
	}
	return AutoStartModeServer
}

// migrateLegacyAutostart 把更名前 "NextPaste" 值名的自启注册项迁移为 "ChuanJian"：
// 保留原启用状态与工作模式，路径更新为当前 exe；开发模式构建跳过，避免写入临时路径。
func migrateLegacyAutostart() {
	execPath, err := getExecPath()
	if err != nil || execPath == "" || isDevBuild(execPath) {
		return
	}

	k, err := registry.OpenKey(registry.CURRENT_USER, autostartRunKey, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		log.Printf("open autostart run key for migration failed: %v", err)
		return
	}
	defer k.Close()

	if err := migrateAutostartValues(k, execPath); err != nil {
		log.Printf("migrate legacy autostart failed: %v", err)
	}
}

func migrateAutostartValues(k registry.Key, execPath string) error {
	legacy, _, err := k.GetStringValue(legacyAutostartValueName)
	if errors.Is(err, registry.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("读取旧自启项失败: %w", err)
	}
	// 先确保新项存在，保留已设置的新配置；写入失败时旧项仍可供重试。
	if _, _, err := k.GetStringValue(autostartValueName); errors.Is(err, registry.ErrNotExist) {
		value := fmt.Sprintf(`"%s" -minimized -autostart=%s`, execPath, parseAutostartMode(legacy))
		if err := k.SetStringValue(autostartValueName, value); err != nil {
			return fmt.Errorf("写入新自启项失败: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("读取新自启项失败: %w", err)
	}
	if err := k.DeleteValue(legacyAutostartValueName); err != nil {
		return fmt.Errorf("删除旧自启项失败: %w", err)
	}
	log.Printf("migrated legacy autostart value %s -> %s", legacyAutostartValueName, autostartValueName)
	return nil
}

// isDevBuild 判断是否运行在开发模式（wails dev）构建产物中：
// wails dev 的二进制位于系统临时目录，指向它的自启注册项无意义
func isDevBuild(execPath string) bool {
	lower := strings.ToLower(execPath)
	tmp := strings.ToLower(os.TempDir())
	return strings.HasPrefix(lower, tmp) || strings.Contains(lower, `\temp\wails\`)
}
