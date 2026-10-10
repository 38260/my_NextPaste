//go:build windows

package main

import (
	_ "embed"
	"log"
	"runtime"
	"unsafe"

	"github.com/getlantern/systray"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/sys/windows"
)

//go:embed build/windows/icon.ico
var trayIcon []byte

// startTray 在独立线程运行托盘消息循环，避免阻塞 Wails 窗口循环
func (a *App) startTray() <-chan struct{} {
	ready := make(chan struct{})
	done := make(chan struct{})
	clicks := make(chan uintptr, 1)
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
				case click := <-clicks:
					a.ShowWindow()
					if click == 0x0205 {
						wailsruntime.EventsEmit(a.ctx, "settings:open")
					}
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

// runTray 将左右键转交给应用，其他消息继续由托盘库处理。
func runTray(onReady, onExit func(), clicks chan<- uintptr) {
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
		// systray v1.2.2 使用 WM_USER+1 通知，区分左键恢复和右键设置。
		if msg.message == 0x0401 && (msg.lParam == 0x0202 || msg.lParam == 0x0205) {
			select {
			case clicks <- msg.lParam:
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
