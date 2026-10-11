//go:build windows

package main

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/windows"
)

func TestBrandMigrationLockedDirectoryRetry(t *testing.T) {
	base := t.TempDir()
	t.Setenv("AppData", base)
	old := filepath.Join(base, legacyWebView2DirNames[0])
	marker := filepath.Join("EBWebView", "saved-config")
	if err := os.MkdirAll(filepath.Join(old, "EBWebView"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(old, marker), []byte("preserved"), 0600); err != nil {
		t.Fatal(err)
	}
	oldPath, err := windows.UTF16PtrFromString(old)
	if err != nil {
		t.Fatal(err)
	}
	// 不共享删除权限，模拟旧实例仍占用目录。
	handle, err := windows.CreateFile(oldPath, windows.GENERIC_READ, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if handle != windows.InvalidHandle {
			windows.CloseHandle(handle)
		}
	})
	if path, err := prepareBrandMigration(); err == nil || path != "" {
		t.Fatalf("占用时应停止启动且不提供新路径，path=%q err=%v", path, err)
	}
	if _, err := os.Stat(userDataPath()); !os.IsNotExist(err) {
		t.Fatalf("失败后不应创建新目录：%v", err)
	}
	if err := windows.CloseHandle(handle); err != nil {
		t.Fatal(err)
	}
	handle = windows.InvalidHandle
	path, err := prepareBrandMigration()
	if err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(filepath.Join(path, marker)); err != nil || string(data) != "preserved" {
		t.Fatalf("重试应保留旧配置，data=%q err=%v", data, err)
	}
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Fatalf("迁移成功后旧路径应已移动：%v", err)
	}
}

func TestWebView2MigrationExistingTarget(t *testing.T) {
	base := t.TempDir()
	t.Setenv("AppData", base)
	for name, value := range map[string]string{brandName: "current", legacyWebView2DirNames[0]: "legacy"} {
		path := filepath.Join(base, name)
		if err := os.Mkdir(path, 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(path, "config"), []byte(value), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := migrateWebView2Data(); err != nil {
		t.Fatal(err)
	}
	for name, value := range map[string]string{brandName: "current", legacyWebView2DirNames[0]: "legacy"} {
		if data, err := os.ReadFile(filepath.Join(base, name, "config")); err != nil || string(data) != value {
			t.Fatalf("已有数据不能覆盖，name=%s data=%q err=%v", name, data, err)
		}
	}
}

func TestWebView2MigrationInvalidPath(t *testing.T) {
	t.Setenv("AppData", "")
	if err := migrateWebView2Data(); err == nil {
		t.Fatal("空 AppData 必须报错，不能使用相对目录")
	}
	base := t.TempDir()
	t.Setenv("AppData", base)
	if err := os.WriteFile(userDataPath(), []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := migrateWebView2Data(); err == nil {
		t.Fatal("目标是文件时必须报错")
	}
	if data, err := os.ReadFile(userDataPath()); err != nil || string(data) != "keep" {
		t.Fatalf("已有文件不能覆盖，data=%q err=%v", data, err)
	}
}
