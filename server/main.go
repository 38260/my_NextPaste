package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	flags := parseFlags()

	// 项目更名为传笺：迁移旧版遗留（WebView2 数据目录、自启注册项），并固定数据目录与 exe 名解耦
	webviewDataPath := prepareBrandMigration()

	// 创建应用实例
	app := NewApp()
	app.SetLaunchFlags(flags)

	// 创建并运行应用
	err := wails.Run(&options.App{
		Title:       "传笺",
		Width:       1200,
		Height:      800,
		MinWidth:    900,
		MinHeight:   600,
		StartHidden: flags.Minimized,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 248, G: 250, B: 252, A: 1}, // 浅色背景 #f8fafc
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		OnBeforeClose:    app.beforeClose,
		// 单实例锁：重复启动时唤起已有窗口，第二实例自行退出
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "ChuanJian-Desktop",
			OnSecondInstanceLaunch: func(secondInstanceData options.SecondInstanceData) {
				app.onSecondInstanceLaunch()
			},
		},
		// WebView2 数据目录与 exe 文件名解耦，exe 改名不再丢失 localStorage 配置
		Windows: &windows.Options{
			WebviewUserDataPath: webviewDataPath,
		},
		// 无边框窗口，使用自定义标题栏
		Frameless: true,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
