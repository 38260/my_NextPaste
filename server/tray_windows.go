//go:build windows

package main

import (
	_ "embed"

	"github.com/getlantern/systray"
)

//go:embed build/windows/icon.ico
var trayIcon []byte

// startTray 在独立线程运行托盘消息循环，避免阻塞 Wails 窗口循环
func (a *App) startTray() <-chan struct{} {
	ready := make(chan struct{})
	done := make(chan struct{})
	go systray.Run(func() {
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
	})
	return ready
}

func (a *App) stopTray() {
	systray.Quit()
	a.trayReady.Store(false)
}
