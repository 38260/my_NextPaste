//go:build !windows

package main

import "log"

// prepareBrandMigration 非 Windows 平台无需迁移旧版遗留，返回空表示用默认数据目录
func prepareBrandMigration() (string, error) {
	return "", nil
}

func showStartupError(err error) {
	log.Printf("startup failed: %v", err)
}
