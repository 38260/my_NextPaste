//go:build windows

package main

import (
	"log"
	"os"
	"path/filepath"
)

// brandName 应用品牌名（更名后），用作 %AppData% 下 WebView2 数据目录名
const brandName = "ChuanJian"

// legacyWebView2DirNames 更名前的 WebView2 数据目录（以旧 exe 文件名命名）
var legacyWebView2DirNames = []string{"NextPaste Server.exe"}

// prepareBrandMigration 项目更名为传笺后的一次性迁移：
// 1) 把旧 exe 名命名的 WebView2 数据目录迁到固定目录，保住 localStorage 里的连接配置与偏好；
// 2) 把旧 "NextPaste" 值名的自启注册项迁移到 "ChuanJian" 并指向新 exe；
// 返回固定后的 WebView2 用户数据目录（与 exe 文件名解耦）。
func prepareBrandMigration() string {
	migrateWebView2Data()
	migrateLegacyAutostart()
	return filepath.Join(os.Getenv("AppData"), brandName)
}

// migrateWebView2Data 迁移旧版 WebView2 数据目录（含 EBWebView），目标已存在则保留现状
func migrateWebView2Data() {
	base := os.Getenv("AppData")
	target := filepath.Join(base, brandName)
	if info, err := os.Stat(target); err == nil && info.IsDir() {
		return // 已迁移或新目录已在使用，保留现状
	}
	for _, legacy := range legacyWebView2DirNames {
		old := filepath.Join(base, legacy)
		info, err := os.Stat(old)
		if err != nil || !info.IsDir() {
			continue
		}
		if err := os.Rename(old, target); err != nil {
			// 旧实例可能仍在运行占用目录，此时放弃迁移，避免影响旧版运行
			log.Printf("migrate WebView2 data dir failed: %v", err)
			return
		}
		log.Printf("migrated WebView2 data dir %s -> %s", old, target)
		return
	}
}

// userDataPath 供测试或诊断使用：返回 WebView2 数据目录
func userDataPath() string {
	return filepath.Join(os.Getenv("AppData"), brandName)
}
