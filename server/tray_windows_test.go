//go:build windows

package main

import (
	"context"
	"os"
	"os/exec"
	"runtime"
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

	ready, done := make(chan struct{}), make(chan struct{})
	clicks := make(chan uintptr, 1)
	shows, quits := make(chan struct{}, 1), make(chan struct{}, 1)
	go runTray(func() {
		systray.SetIcon(trayIcon)
		show := systray.AddMenuItem("显示窗口", "")
		systray.AddSeparator()
		quit := systray.AddMenuItem("退出应用", "")
		go func() {
			for {
				select {
				case <-show.ClickedCh:
					shows <- struct{}{}
				case <-quit.ClickedCh:
					quits <- struct{}{}
				case <-done:
					return
				}
			}
		}()
		close(ready)
	}, func() { close(done) }, clicks)
	t.Cleanup(systray.Quit)
	wait := func(ch <-chan struct{}, event string) {
		t.Helper()
		select {
		case <-ch:
		case <-time.After(3 * time.Second):
			t.Fatalf("timeout waiting for %s", event)
		}
	}
	wait(ready, "tray ready")
	waitClick := func(expected uintptr) {
		t.Helper()
		select {
		case actual := <-clicks:
			if actual != expected {
				t.Fatalf("tray click = %#x, want %#x", actual, expected)
			}
		case <-time.After(3 * time.Second):
			t.Fatalf("timeout waiting for tray click %#x", expected)
		}
	}

	user32 := windows.NewLazySystemDLL("user32.dll")
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

	// 重复调度与垃圾回收后，真实窗口队列仍须处理每次左键通知。
	for i := 0; i < 25; i++ {
		runtime.GC()
		runtime.Gosched()
		post(0x0401, 100, 0x0202)
		waitClick(0x0202)
	}

	for i := 0; i < 3; i++ {
		post(0x0401, 100, 0x0205)
		waitClick(0x0205)
		menuVisible := false
		user32.NewProc("EnumThreadWindows").Call(thread, windows.NewCallback(func(hwnd, _ uintptr) uintptr {
			var name [64]uint16
			user32.NewProc("GetClassNameW").Call(hwnd, uintptr(unsafe.Pointer(&name[0])), uintptr(len(name)))
			visible, _, _ := user32.NewProc("IsWindowVisible").Call(hwnd)
			if windows.UTF16ToString(name[:]) == "#32768" && visible != 0 {
				menuVisible = true
			}
			return 1
		}), 0)
		if menuVisible {
			t.Fatal("right click opened the old menu instead of settings")
		}
		post(0x0111, 1, 0) // WM_COMMAND：库分配的第一个菜单项为显示窗口。
		wait(shows, "show menu action")
		select {
		case <-clicks:
			t.Fatal("right click incorrectly triggered left click")
		default:
		}
	}
	post(0x0111, 3, 0)
	wait(quits, "quit menu action")
	systray.Quit()
	wait(done, "tray shutdown")
}
