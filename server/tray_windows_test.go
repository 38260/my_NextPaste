//go:build windows

package main

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/getlantern/systray"
	"golang.org/x/sys/windows"
)

func TestTrayEvents(t *testing.T) {
	// 托盘库有进程级状态，用子进程隔离，避免影响其他测试及用户实例。
	if os.Getenv("NEXT_PASTE_TRAY_TEST") != "1" {
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestTrayEvents$", "-test.v")
		cmd.Env = append(os.Environ(), "NEXT_PASTE_TRAY_TEST=1")
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("tray subprocess: %v\n%s", err, output)
		}
		return
	}

	app := NewApp()
	ready, done := make(chan struct{}), make(chan struct{})
	clicks := make(chan uintptr, 1)
	rights := make(chan struct{}, 1)
	commands := make(chan uintptr, 8)
	go runTray(func() {
		systray.SetIcon(trayIcon)
		close(ready)
	}, func() { close(done) }, clicks, func(hwnd uintptr) {
		app.showTrayMenu(hwnd)
		rights <- struct{}{}
	}, func(id uintptr) { commands <- id })
	t.Cleanup(systray.Quit)
	wait := func(ch <-chan struct{}, event string) {
		t.Helper()
		select {
		case <-ch:
		case <-time.After(5 * time.Second):
			t.Fatalf("timeout waiting for %s", event)
		}
	}
	waitValue := func(ch <-chan uintptr, event string) uintptr {
		t.Helper()
		select {
		case value := <-ch:
			return value
		case <-time.After(5 * time.Second):
			t.Fatalf("timeout waiting for %s", event)
			return 0
		}
	}
	waitClick := func(expected uintptr) {
		t.Helper()
		select {
		case actual := <-clicks:
			if actual != expected {
				t.Fatalf("tray click = %#x, want %#x", actual, expected)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("timeout waiting for tray click %#x", expected)
		}
	}

	user32 := windows.NewLazySystemDLL("user32.dll")
	wait(ready, "tray ready")
	var tray, thread uintptr
	user32.NewProc("EnumWindows").Call(windows.NewCallback(func(hwnd, _ uintptr) uintptr {
		var pid uint32
		tid, _, _ := user32.NewProc("GetWindowThreadProcessId").Call(hwnd, uintptr(unsafe.Pointer(&pid)))
		var name [64]uint16
		user32.NewProc("GetClassNameW").Call(hwnd, uintptr(unsafe.Pointer(&name[0])), uintptr(len(name)))
		if pid == uint32(os.Getpid()) && windows.UTF16ToString(name[:]) == "SystrayClass" {
			tray, thread = hwnd, tid
			return 0
		}
		return 1
	}), 0)
	if tray == 0 {
		t.Fatal("tray window not found")
	}
	post := func(message, wParam, lParam uintptr) {
		t.Helper()
		if result, _, err := user32.NewProc("PostMessageW").Call(tray, message, wParam, lParam); result == 0 {
			t.Fatalf("post tray message: %v", err)
		}
	}
	menuVisible := func() bool {
		t.Helper()
		visible := false
		user32.NewProc("EnumThreadWindows").Call(thread, windows.NewCallback(func(hwnd, _ uintptr) uintptr {
			var name [64]uint16
			user32.NewProc("GetClassNameW").Call(hwnd, uintptr(unsafe.Pointer(&name[0])), uintptr(len(name)))
			shown, _, _ := user32.NewProc("IsWindowVisible").Call(hwnd)
			if windows.UTF16ToString(name[:]) == "#32768" && shown != 0 {
				visible = true
			}
			return 1
		}), 0)
		return visible
	}
	waitMenu := func(want bool, what string) {
		t.Helper()
		deadline := time.Now().Add(3 * time.Second)
		for time.Now().Before(deadline) {
			if menuVisible() == want {
				return
			}
			time.Sleep(50 * time.Millisecond)
		}
		t.Fatalf("menu visible = %v, want %v (%s)", !want, want, what)
	}

	// 重复调度与垃圾回收后，真实窗口队列仍须处理每次左键通知。
	for i := 0; i < 25; i++ {
		runtime.GC()
		runtime.Gosched()
		post(0x0401, 100, 0x0202)
		waitClick(0x0202)
	}

	// 左键不应弹出任何菜单。
	post(0x0401, 100, 0x0202)
	waitClick(0x0202)
	time.Sleep(300 * time.Millisecond)
	if menuVisible() {
		t.Fatal("left click opened a menu")
	}

	// 右键应弹出最近记录菜单，取消后应正常收起。
	post(0x0401, 100, 0x0205)
	waitMenu(true, "right click should open the record menu")
	post(0x001F, 0, 0) // WM_CANCELMODE：收起菜单
	wait(rights, "tray menu closed")
	waitMenu(false, "menu should close after WM_CANCELMODE")

	// 自绘菜单命令在消息循环拦截并转交，其余命令继续走托盘库。
	post(0x0111, trayMenuSettings, 0)
	if got := waitValue(commands, "settings command"); got != trayMenuSettings {
		t.Fatalf("settings command = %#x, want %#x", got, trayMenuSettings)
	}
	post(0x0111, trayMenuRecordBase+2, 0)
	if got := waitValue(commands, "record command"); got != trayMenuRecordBase+2 {
		t.Fatalf("record command = %#x, want %#x", got, trayMenuRecordBase+2)
	}
	post(0x0111, 1, 0) // 托盘库内部菜单 ID，不应进入自绘命令通道。
	select {
	case got := <-commands:
		t.Fatalf("non-custom command %#x incorrectly routed", got)
	case <-time.After(500 * time.Millisecond):
	}

	// 再次右键弹出后正常收起，保证菜单可重复打开。
	post(0x0401, 100, 0x0205)
	waitMenu(true, "reopened record menu")
	post(0x001F, 0, 0)
	wait(rights, "tray menu closed again")
	waitMenu(false, "menu should close again")

	systray.Quit()
	wait(done, "tray shutdown")
}

func TestRecentClipRecords(t *testing.T) {
	app := NewApp()
	if got := app.recentClipRecords(trayRecentCount); len(got) != 0 {
		t.Fatalf("empty logs should yield no records, got %d", len(got))
	}

	app.onLogClip("INFO", "检测到剪贴板文本变化: 6 字符", "text", []byte("第一条内容"))
	app.onLogClip("INFO", "检测到剪贴板图片变化: 1.00 MB", "image", make([]byte, 1024*1024))
	app.onLogClip("INFO", "检测到剪贴板文本变化: 6 字符", "text", []byte("第二条内容"))
	app.onLog("INFO", "系统日志不进托盘菜单")

	records := app.recentClipRecords(trayRecentCount)
	if len(records) != 3 {
		t.Fatalf("recent records = %d, want 3", len(records))
	}
	if !strings.Contains(records[0].Label, "第二条内容") {
		t.Errorf("newest record should come first, label = %q", records[0].Label)
	}
	if !strings.Contains(records[1].Label, "图片 · 1.00 MB") {
		t.Errorf("image record label = %q, want size summary", records[1].Label)
	}
	for _, record := range records {
		if record.ID == 0 {
			t.Errorf("record ID should reference stored content, label = %q", record.Label)
		}
		if !strings.Contains(record.Tooltip, " · ") {
			t.Errorf("tooltip should carry time prefix, tooltip = %q", record.Tooltip)
		}
	}

	// 超长文本在菜单标题里截断并带省略号，完整预览留在悬停提示。
	long := strings.Repeat("长", 60)
	app.onLogClip("INFO", "检测到剪贴板文本变化", "text", []byte(long))
	records = app.recentClipRecords(trayRecentCount)
	if !strings.Contains(records[0].Label, "…") || len([]rune(records[0].Label)) > 60 {
		t.Errorf("long text should be truncated with ellipsis, label = %q", records[0].Label)
	}
	if !strings.Contains(records[0].Tooltip, long) {
		t.Errorf("tooltip should keep the full preview, tooltip = %q", records[0].Tooltip)
	}

	// 清空后菜单数据同步清空。
	if err := app.ClearLogs(); err != nil {
		t.Fatalf("clear logs: %v", err)
	}
	if got := app.recentClipRecords(trayRecentCount); len(got) != 0 {
		t.Fatalf("cleared logs should yield no records, got %d", len(got))
	}
}
