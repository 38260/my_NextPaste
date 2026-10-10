//go:build !windows

package main

// 非 Windows 平台暂不提供托盘，窗口关闭时保留任务栏入口
func (a *App) startTray() <-chan struct{} {
	ready := make(chan struct{})
	close(ready)
	return ready
}

func (a *App) stopTray() {}
