//go:build windows

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"unsafe"

	"golang.org/x/sys/windows"
)

// brandName 应用品牌名（更名后），用作 %AppData% 下 WebView2 数据目录名
const brandName = "ChuanJian"

// legacyWebView2DirNames 更名前的 WebView2 数据目录（以旧 exe 文件名命名）
var legacyWebView2DirNames = []string{"NextPaste Server.exe"}

// prepareBrandMigration 项目更名为传笺后的一次性迁移：
// 1) 把旧 exe 名命名的 WebView2 数据目录迁到固定目录，保住 localStorage 里的连接配置与偏好；
// 2) 把旧 "NextPaste" 值名的自启注册项迁移到 "ChuanJian" 并指向新 exe；
// 返回固定后的 WebView2 用户数据目录（与 exe 文件名解耦）。
func prepareBrandMigration() (string, error) {
	if err := migrateWebView2Data(); err != nil {
		return "", err
	}
	migrateLegacyAutostart()
	return userDataPath(), nil
}

// migrateWebView2Data 迁移旧版 WebView2 数据目录（含 EBWebView），目标已存在则保留现状
func migrateWebView2Data() error {
	base := os.Getenv("AppData")
	if base == "" {
		return errors.New("AppData 目录为空，无法确定用户数据位置")
	}
	target := filepath.Join(base, brandName)
	if info, err := os.Stat(target); err == nil {
		if !info.IsDir() {
			return fmt.Errorf("用户数据位置不是目录：%s", target)
		}
		return nil // 已迁移或新目录已在使用，保留现状
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("检查用户数据目录失败：%w", err)
	}
	for _, legacy := range legacyWebView2DirNames {
		old := filepath.Join(base, legacy)
		info, err := os.Stat(old)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("检查旧版用户数据目录失败：%w", err)
		}
		if !info.IsDir() {
			continue
		}
		if err := os.Rename(old, target); err != nil {
			// 迁移失败时阻止 WebView2 初始化新目录，保留下次重试的机会。
			return fmt.Errorf("迁移旧版用户数据失败：%w", err)
		}
		log.Printf("migrated WebView2 data dir %s -> %s", old, target)
		return nil
	}
	return nil
}

// 无控制台的正式程序也必须显示启动失败原因。
func showStartupError(err error) {
	log.Printf("startup failed: %v", err)
	message, _ := windows.UTF16PtrFromString("无法启动传笺：" + err.Error() + "\n请关闭旧版程序或检查数据目录权限，然后重试。")
	title, _ := windows.UTF16PtrFromString("传笺 · 启动失败")
	windows.NewLazySystemDLL("user32.dll").NewProc("MessageBoxW").Call(0, uintptr(unsafe.Pointer(message)), uintptr(unsafe.Pointer(title)), 0x10)
}

// userDataPath 供测试或诊断使用：返回 WebView2 数据目录
func userDataPath() string {
	return filepath.Join(os.Getenv("AppData"), brandName)
}
