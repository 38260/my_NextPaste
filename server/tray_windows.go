//go:build windows

package main

import (
	_ "embed"
	"log"
	"runtime"
	"unsafe"

	"github.com/getlantern/systray"
	"golang.org/x/sys/windows"
)

//go:embed build/windows/icon.ico
var trayIcon []byte

// startTray 在独立线程运行托盘消息循环，避免阻塞 Wails 窗口循环
func (a *App) startTray() <-chan struct{} {
	ready := make(chan struct{})
	done := make(chan struct{})
	clicks := make(chan struct{}, 1)
	go runTray(func() {
		systray.SetIcon(trayIcon)
		systray.SetTooltip("传笺 · 后台同步")
		show := systray.AddMenuItem("显示窗口", "恢复传笺主窗口")
		systray.AddSeparator()
		quit := systray.AddMenuItem("退出应用", "停止同步并退出传笺")
		a.trayReady.Store(true)
		close(ready)

		go func() {
			for {
				select {
				case <-clicks:
					a.ShowWindow()
				case <-show.ClickedCh:
					a.ShowWindow()
				case <-quit.ClickedCh:
					a.Quit()
				case <-done:
					return
				}
			}
		}()
	}, func() {
		a.trayReady.Store(false)
		close(done)
	}, clicks)
	return ready
}

// runTray 保留库的菜单处理，仅将左键消息转交给窗口恢复逻辑。
func runTray(onReady, onExit func(), clicks chan<- struct{}) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer onExit()
	systray.Register(onReady, nil)

	user32 := windows.NewLazySystemDLL("user32.dll")
	getMessage := user32.NewProc("GetMessageW")
	translateMessage := user32.NewProc("TranslateMessage")
	dispatchMessage := user32.NewProc("DispatchMessageW")
	var msg struct {
		window  uintptr
		message uint32
		wParam  uintptr
		lParam  uintptr
		time    uint32
		x, y    int32
		private uint32
	}
	for {
		result, _, err := getMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(result) <= 0 {
			if int32(result) == -1 {
				log.Printf("tray message loop failed: %v", err)
				systray.Quit()
			}
			return
		}
		// systray v1.2.2 使用 WM_USER+1 通知；右键及其他消息继续交给库。
		if msg.message == 0x0401 && msg.lParam == 0x0202 {
			select {
			case clicks <- struct{}{}:
			default:
			}
			continue
		}
		translateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		dispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

func (a *App) stopTray() {
	systray.Quit()
	a.trayReady.Store(false)
}
