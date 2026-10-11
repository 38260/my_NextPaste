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

// 托盘右键自绘菜单的命令 ID 与相关 Win32 常量。
// 命令 ID 从 0x7000 起，避免与托盘库内部菜单项冲突。
const (
	trayMenuRecordBase = 0x7000 // 最近记录条目：base + 下标
	trayMenuSettings   = 0x7100 // 打开设置
	trayMenuQuit       = 0x7101 // 退出应用

	wmTrayNotify = 0x0401 // systray v1.2.2 的托盘通知消息（WM_USER+1）
	wmCommand    = 0x0111
	wmLButtonUp  = 0x0202
	wmRButtonUp  = 0x0205
	wmCancelMode = 0x001F

	mfString       = 0x0000
	mfSeparator    = 0x0800
	tpmLeftAlign   = 0x0000
	tpmBottomAlign = 0x0020
)

var (
	trayUser32          = windows.NewLazySystemDLL("user32.dll")
	procGetCursorPos    = trayUser32.NewProc("GetCursorPos")
	procSetForeground   = trayUser32.NewProc("SetForegroundWindow")
	procCreatePopupMenu = trayUser32.NewProc("CreatePopupMenu")
	procAppendMenuW     = trayUser32.NewProc("AppendMenuW")
	procTrackPopupMenu  = trayUser32.NewProc("TrackPopupMenu")
	procDestroyMenu     = trayUser32.NewProc("DestroyMenu")
	procPostMessageW    = trayUser32.NewProc("PostMessageW")
)

// startTray 在独立线程运行托盘消息循环，避免阻塞 Wails 窗口循环
func (a *App) startTray() <-chan struct{} {
	ready := make(chan struct{})
	done := make(chan struct{})
	clicks := make(chan uintptr, 1)
	go runTray(func() {
		systray.SetIcon(trayIcon)
		systray.SetTooltip("传笺 · 后台同步")
		a.trayReady.Store(true)
		close(ready)
	}, func() {
		a.trayReady.Store(false)
		close(done)
	}, clicks, a.showTrayMenu, a.routeTrayCommand)

	go func() {
		for {
			select {
			case <-clicks:
				// 左键：直接恢复主界面，前端收到 window:restored 后收起设置面板。
				a.ShowWindow()
				wailsruntime.EventsEmit(a.ctx, "window:restored")
			case <-done:
				return
			}
		}
	}()
	return ready
}

// showTrayMenu 在托盘窗口弹出右键菜单（同步阻塞至菜单关闭）。
// 每次弹出前按当前记录重建菜单，清空、淘汰后自然及时更新。
func (a *App) showTrayMenu(hwnd uintptr) {
	records := a.recentClipRecords(trayRecentCount)
	a.trayMenuMu.Lock()
	a.trayMenuRecords = records
	a.trayMenuMu.Unlock()

	menu, _, _ := procCreatePopupMenu.Call()
	if menu == 0 {
		return
	}
	defer procDestroyMenu.Call(menu)

	appendItem := func(id uintptr, label string) {
		text, err := windows.UTF16PtrFromString(label)
		if err != nil {
			return
		}
		procAppendMenuW.Call(menu, mfString, id, uintptr(unsafe.Pointer(text)))
	}
	for i, record := range records {
		appendItem(trayMenuRecordBase+uintptr(i), record.Label)
	}
	if len(records) > 0 {
		procAppendMenuW.Call(menu, mfSeparator, 0, 0)
	}
	appendItem(trayMenuSettings, "打开设置")
	appendItem(trayMenuQuit, "退出应用")

	var point struct{ X, Y int32 }
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&point)))
	// MSDN：托盘弹出菜单需先把宿主窗口设为前台，点击菜单外部才能正常收起；
	// 菜单关闭后补发 WM_NULL 让前台状态复位。
	procSetForeground.Call(hwnd)
	procTrackPopupMenu.Call(menu, tpmLeftAlign|tpmBottomAlign, uintptr(point.X), uintptr(point.Y), 0, hwnd, 0)
	procPostMessageW.Call(hwnd, 0, 0, 0)
}

// routeTrayCommand 处理自绘菜单命令（在托盘消息循环线程调用）
func (a *App) routeTrayCommand(id uintptr) {
	switch {
	case id == trayMenuSettings:
		a.showSettings()
	case id == trayMenuQuit:
		a.Quit()
	case id >= trayMenuRecordBase && id < trayMenuRecordBase+trayRecentCount:
		a.trayMenuMu.Lock()
		var target int64
		if index := int(id - trayMenuRecordBase); index < len(a.trayMenuRecords) {
			target = a.trayMenuRecords[index].ID
		}
		a.trayMenuMu.Unlock()
		if target != 0 {
			a.showRecord(target)
		}
	}
}

// runTray 将左键通知、右键通知与自绘菜单命令转交给应用，其余消息继续由托盘库处理。
func runTray(onReady, onExit func(), clicks chan<- uintptr, onRightClick func(hwnd uintptr), onCommand func(id uintptr)) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer onExit()
	systray.Register(onReady, nil)

	getMessage := trayUser32.NewProc("GetMessageW")
	translateMessage := trayUser32.NewProc("TranslateMessage")
	dispatchMessage := trayUser32.NewProc("DispatchMessageW")
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
		// systray v1.2.2 使用 WM_USER+1 通知：左键转交应用恢复窗口，右键弹出记录菜单。
		if msg.message == wmTrayNotify {
			switch msg.lParam {
			case wmLButtonUp:
				select {
				case clicks <- msg.lParam:
				default:
				}
				continue
			case wmRButtonUp:
				if onRightClick != nil {
					onRightClick(msg.window)
				}
				continue
			}
		}
		if msg.message == wmCommand && onCommand != nil && msg.wParam&0xFFFF >= trayMenuRecordBase {
			onCommand(msg.wParam & 0xFFFF)
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
