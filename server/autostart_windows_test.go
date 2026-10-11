//go:build windows

package main

import "testing"

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
