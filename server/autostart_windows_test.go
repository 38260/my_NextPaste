//go:build windows

package main

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/windows/registry"
)

func TestAutostartMigration(t *testing.T) {
	const currentPath = `D:\app\ChuanJian.exe`
	const legacy = `"D:\app\NextPaste Server.exe" -minimized -autostart=client`
	const current = `"D:\custom\ChuanJian.exe" -minimized -autostart=server`
	for _, test := range []struct {
		name      string
		legacy    string
		current   string
		readOnly  bool
		wantError string
		wantMode  string
	}{
		{name: "absent"},
		{name: "migrate-client", legacy: legacy, wantMode: "client"},
		{name: "migrate-server", legacy: `"old.exe" -minimized`, wantMode: "server"},
		{name: "keep-current", legacy: legacy, current: current},
		{name: "write-failure-retry", legacy: legacy, readOnly: true, wantError: "写入新自启项失败", wantMode: "client"},
		{name: "delete-failure-retry", legacy: legacy, current: current, readOnly: true, wantError: "删除旧自启项失败"},
	} {
		t.Run(test.name, func(t *testing.T) {
			// 独立测试键，不读写用户 Run 键；只读句柄制造真实写入/删除失败。
			path := `Software\ChuanJianMigrationTest-` + filepath.Base(filepath.Dir(t.TempDir()))
			key, existed, err := registry.CreateKey(registry.CURRENT_USER, path, registry.QUERY_VALUE|registry.SET_VALUE)
			if err != nil {
				t.Fatal(err)
			}
			if existed {
				key.Close()
				t.Fatal("测试键已存在，拒绝修改")
			}
			t.Cleanup(func() {
				key.Close()
				if err := registry.DeleteKey(registry.CURRENT_USER, path); err != nil {
					t.Errorf("清理测试键失败：%v", err)
				}
			})
			for name, value := range map[string]string{legacyAutostartValueName: test.legacy, autostartValueName: test.current} {
				if value != "" {
					if err := key.SetStringValue(name, value); err != nil {
						t.Fatal(err)
					}
				}
			}
			migrationKey := key
			if test.readOnly {
				migrationKey, err = registry.OpenKey(registry.CURRENT_USER, path, registry.QUERY_VALUE)
				if err != nil {
					t.Fatal(err)
				}
				defer migrationKey.Close()
			}
			err = migrateAutostartValues(migrationKey, currentPath)
			if test.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantError) {
					t.Fatalf("预期 %s，实际 %v", test.wantError, err)
				}
				if value, _, err := key.GetStringValue(legacyAutostartValueName); err != nil || value != test.legacy {
					t.Fatalf("失败时旧项必须保留，value=%q err=%v", value, err)
				}
				if test.current == "" {
					if _, _, err := key.GetStringValue(autostartValueName); !errors.Is(err, registry.ErrNotExist) {
						t.Fatalf("写入失败不应产生新项：%v", err)
					}
				}
				if err := migrateAutostartValues(key, currentPath); err != nil {
					t.Fatalf("恢复写权限后重试失败：%v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if _, _, err := key.GetStringValue(legacyAutostartValueName); !errors.Is(err, registry.ErrNotExist) {
				t.Fatalf("完成后旧项应已删除：%v", err)
			}
			value, _, err := key.GetStringValue(autostartValueName)
			if test.legacy == "" {
				if !errors.Is(err, registry.ErrNotExist) {
					t.Fatalf("没有旧项不应启用自启：%v", err)
				}
				return
			}
			want := test.current
			if want == "" {
				want = `"` + currentPath + `" -minimized -autostart=` + test.wantMode
			}
			if err != nil || value != want {
				t.Fatalf("新项 value=%q want=%q err=%v", value, want, err)
			}
		})
	}
}

// TestAutostartDevGuard 验证开发模式防护：go test 的二进制位于系统临时目录，
// setAutostart(true) 必须被 isDevBuild 拦截并返回明确错误，且不产生任何注册表写入
func TestAutostartDevGuard(t *testing.T) {
	execPath, err := getExecPath()
	if err != nil {
		t.Fatalf("获取测试进程路径失败: %v", err)
	}
	if !isDevBuild(execPath) {
		t.Fatalf("测试二进制应位于临时目录以触发 dev 防护，实际路径: %s", execPath)
	}

	if err := setAutostart(true, AutoStartModeServer); err == nil {
		t.Fatal("开发模式下开启自启应返回错误，实际返回 nil（可能已写入注册表，请人工检查 Run 键）")
	}

	// 只读查询不应报错；此测试不执行删除，避免误删用户真实注册项
	enabled, _, err := isAutostartEnabled()
	if err != nil {
		t.Fatalf("查询自启状态不应报错: %v", err)
	}
	if enabled {
		t.Log("注意：检测到已存在的自启注册项（本测试未写入，应为环境遗留）")
	}
}
