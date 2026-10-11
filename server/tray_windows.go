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
	trayMenuAutostart  = 0x7102 // 开机自启开关

	wmTrayNotify = 0x0401 // systray v1.2.2 的托盘通知消息（WM_USER+1）
	wmCommand    = 0x0111
	wmLButtonUp  = 0x0202
	wmRButtonUp  = 0x0205
	wmCancelMode = 0x001F

	mfString       = 0x0000
	mfSeparator    = 0x0800
	mfChecked      = 0x0008
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

	appendItem := func(id uintptr, label string, flags uintptr) {
		text, err := windows.UTF16PtrFromString(label)
		if err != nil {
			return
		}
		procAppendMenuW.Call(menu, flags, id, uintptr(unsafe.Pointer(text)))
	}
	for i, record := range records {
		appendItem(trayMenuRecordBase+uintptr(i), record.Label, mfString)
	}
	if len(records) > 0 {
		procAppendMenuW.Call(menu, mfSeparator, 0, 0)
	}
	autostartFlags := uintptr(mfString)
	if a.autostartEnabled() {
		autostartFlags |= mfChecked
	}
	appendItem(trayMenuAutostart, "开机自启", autostartFlags)
	appendItem(trayMenuSettings, "打开设置", mfString)
	appendItem(trayMenuQuit, "退出应用", mfString)

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
	case id == trayMenuAutostart:
		a.toggleTrayAutostart()
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
	systray.Register(nil, nil)

	var trayWindow uintptr
	trayUser32.NewProc("EnumThreadWindows").Call(uintptr(windows.GetCurrentThreadId()), windows.NewCallback(func(hwnd, _ uintptr) uintptr {
		var name [64]uint16
		trayUser32.NewProc("GetClassNameW").Call(hwnd, uintptr(unsafe.Pointer(&name[0])), uintptr(len(name)))
		if windows.UTF16ToString(name[:]) == "SystrayClass" {
			trayWindow = hwnd
			return 0
		}
		return 1
	}), 0)
	if trayWindow == 0 {
		log.Print("tray window not found on its message-loop thread")
		systray.Quit()
		return
	}
	controls := windows.NewLazySystemDLL("comctl32.dll")
	defSubclass := controls.NewProc("DefSubclassProc")
	removeSubclass := controls.NewProc("RemoveWindowSubclass")
	var callback uintptr
	callback = windows.NewCallback(func(hwnd uintptr, message uint32, wParam, lParam, _id, _data uintptr) uintptr {
		// Explorer 的直接通知和队列通知都经过窗口过程，不能只在 GetMessage 后拦截。
		if message == wmTrayNotify {
			switch lParam {
			case wmLButtonUp:
				select {
				case clicks <- lParam:
				default:
				}
				return 0
			case wmRButtonUp:
				if onRightClick != nil {
					onRightClick(hwnd)
				}
				return 0
			}
		}
		if message == wmCommand && onCommand != nil && wParam&0xFFFF >= trayMenuRecordBase {
			onCommand(wParam & 0xFFFF)
			return 0
		}
		if message == 0x0082 { // WM_NCDESTROY
			removeSubclass.Call(hwnd, callback, 1)
		}
		result, _, _ := defSubclass.Call(hwnd, uintptr(message), wParam, lParam)
		return result
	})
	if ok, _, err := controls.NewProc("SetWindowSubclass").Call(trayWindow, callback, 1, 0); ok == 0 {
		log.Printf("install tray window callback failed: %v", err)
		systray.Quit()
		return
	}
	onReady()

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
		translateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		dispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

func (a *App) stopTray() {
	systray.Quit()
	a.trayReady.Store(false)
}
