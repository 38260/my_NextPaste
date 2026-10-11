//go:build !windows

package main

// prepareBrandMigration 非 Windows 平台无需迁移旧版遗留，返回空表示用默认数据目录
func prepareBrandMigration() string {
	return ""
}
