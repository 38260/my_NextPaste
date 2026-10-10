package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"server/internal/clipboard"
	ws "server/internal/websocket"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// LogEntry 日志条目
type LogEntry struct {
	Level     string `json:"level"`
	Message   string `json:"message"`
	Preview   string `json:"preview"`   // 剪贴板内容预览（非剪贴板日志为空）
	Category  string `json:"category"`  // "clipboard" | "connection" | "system"
	ContentID int64  `json:"contentId"` // 关联的暂存内容 ID（0 表示无）
	Timestamp int64  `json:"timestamp"`
}

// storedContent 暂存的剪贴板内容（用于日志缩略图展示与一键复制）
type storedContent struct {
	dataType string
	content  []byte
}

const (
	maxStoredContents   = 20
	maxStoredImageBytes = 20 * 1024 * 1024 // 图片超过 20MB 不暂存（日志仍正常记录）
)

// App struct（V1.1 二进制协议版本）
type App struct {
	ctx          context.Context
	wsServer     *ws.Server
	wsClient     *ws.WSClient
	clipboardMon *clipboard.Monitor
	logs         []LogEntry
	logsMu       sync.RWMutex
	maxLogs      int
	mode         string // "server" 或 "client"

	settingsMu sync.RWMutex
	settings   Settings

	contents      map[int64]storedContent
	contentsOrder []int64
	contentsMu    sync.RWMutex
	contentNext   int64

	clip *clipLedger

	launchFlags LaunchFlags
	trayReady   atomic.Bool
	quitting    atomic.Bool
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		wsServer:     ws.NewServer(),
		wsClient:     ws.NewWSClient("NextPaste Desktop", "Windows"),
		clipboardMon: clipboard.NewMonitor(),
		logs:         make([]LogEntry, 0),
		maxLogs:      500,
		contents:     make(map[int64]storedContent),
		clip:         newClipLedger(),
		mode:         "server", // 默认为服务器模式
		settings: Settings{
			AutoStartEnabled:       false,
			AutoStartActionEnabled: false,
			AutoStartMode:          AutoStartModeServer,
			CloseAction:            "tray",
		},
		launchFlags: LaunchFlags{},
	}
}

func (a *App) SetLaunchFlags(flags LaunchFlags) {
	a.launchFlags = flags
}

// ============================================
// 窗口控制方法
// ============================================

// ShowWindow 显示主窗口
func (a *App) ShowWindow() {
	runtime.WindowShow(a.ctx)
	runtime.WindowUnminimise(a.ctx)
}

// HideWindow 将主窗口隐藏到托盘；托盘尚未就绪时保留任务栏入口
func (a *App) HideWindow() {
	if a.trayReady.Load() {
		runtime.WindowHide(a.ctx)
		return
	}
	runtime.WindowMinimise(a.ctx)
}

// Quit 完全退出应用
func (a *App) Quit() {
	a.quitting.Store(true)
	runtime.Quit(a.ctx)
}

// ApplySettings updates settings and applies side effects (autostart).
func (a *App) ApplySettings(settingsJSON string) error {
	var s Settings
	if err := json.Unmarshal([]byte(settingsJSON), &s); err != nil {
		return err
	}

	// Normalize
	if s.AutoStartMode != AutoStartModeClient {
		s.AutoStartMode = AutoStartModeServer
	}
	if s.CloseAction != "close" {
		s.CloseAction = "tray"
	}
	// 关闭偏好必须先生效；开机自启写入失败不应让窗口行为退回默认值
	a.settingsMu.Lock()
	a.settings.CloseAction = s.CloseAction
	a.settingsMu.Unlock()

	// OS 级开机自启：Windows 走注册表 Run 键（纯 Go），
	// macOS/Linux 沿用 go-autostart，见 autostart_windows.go / autostart_other.go
	if err := setAutostart(s.AutoStartEnabled, s.AutoStartMode); err != nil {
		// 开启失败不吞掉：让前端弹错并把开关回滚为注册表真实状态
		return err
	}

	a.settingsMu.Lock()
	a.settings = s
	a.settingsMu.Unlock()
	return nil
}

// GetSettings returns current settings.
func (a *App) GetSettings() map[string]any {
	a.settingsMu.RLock()
	defer a.settingsMu.RUnlock()

	return map[string]any{
		"autoStartEnabled":       a.settings.AutoStartEnabled,
		"autoStartActionEnabled": a.settings.AutoStartActionEnabled,
		"autoStartMode":          string(a.settings.AutoStartMode),
		"closeAction":            a.settings.CloseAction,
	}
}

// GetAutostartStatus 查询 OS 级开机自启的真实注册状态（以系统为准，而非本地设置副本）
func (a *App) GetAutostartStatus() map[string]any {
	enabled, mode, err := isAutostartEnabled()
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	return map[string]any{
		"enabled": enabled,
		"mode":    string(mode),
		"error":   errMsg,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	trayReady := a.startTray()

	// Settings live in localStorage on the frontend; backend stores the last applied copy.
	// Auto start/connect is triggered from the frontend (so it can reuse persisted
	// address/port/url and show UI errors).
	if a.launchFlags.Minimized {
		select {
		case <-trayReady:
		case <-time.After(5 * time.Second):
		}
		a.HideWindow()
	}
}

// beforeClose 拦截关窗和退出请求；显式退出由 quitting 标志放行
func (a *App) beforeClose(ctx context.Context) bool {
	if a.quitting.Load() {
		return false
	}
	a.settingsMu.RLock()
	action := a.settings.CloseAction
	a.settingsMu.RUnlock()

	if action != "close" {
		a.HideWindow()
		return true
	}
	return false
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	a.stopTray()
	a.StopServer()
	a.DisconnectClient()
}

// StartServer 启动 WebSocket 服务器
func (a *App) StartServer(address string, port int) error {
	if a.wsServer.IsRunning() {
		return fmt.Errorf("服务器已在运行中")
	}

	// 设置剪贴板数据接收回调（V1.1 二进制协议）
	a.wsServer.SetClipboardCallback(a.onClipboardReceivedBinary)

	// 启动 WebSocket 服务器
	err := a.wsServer.Start(address, port, a.onLogFull)
	if err != nil {
		return err
	}

	// 启动剪贴板监听
	err = a.clipboardMon.Start(a.onClipboardChange)
	if err != nil {
		a.wsServer.Stop()
		return err
	}

	return nil
}

// StopServer 停止 WebSocket 服务器
func (a *App) StopServer() error {
	if a.mode == "server" {
		a.clipboardMon.Stop()
	}
	return a.wsServer.Stop()
}

// ConnectClient 连接到远程 WebSocket 服务器（客户端模式）
func (a *App) ConnectClient(url string) error {
	if a.wsClient.IsConnected() {
		return fmt.Errorf("客户端已连接")
	}

	a.mode = "client"

	// 设置剪贴板数据接收回调（V1.1 二进制协议）
	a.wsClient.SetClipboardCallback(a.onClipboardReceivedBinary)

	// 设置连接成功回调 - 只有连接成功后才启动剪贴板监听
	a.wsClient.SetOnConnected(func() {
		a.onLog("INFO", "WebSocket 连接成功，启动剪贴板监听...")
		err := a.clipboardMon.Start(a.onClipboardChangeClient)
		if err != nil {
			a.onLog("ERROR", fmt.Sprintf("启动剪贴板监听失败: %v", err))
		} else {
			a.onLog("SUCCESS", "剪贴板监听已启动")
		}
	})

	// 连接到服务器（异步，会自动重连）
	err := a.wsClient.Connect(url, a.onLogFull)
	if err != nil {
		return err
	}

	return nil
}

// DisconnectClient 断开客户端连接
func (a *App) DisconnectClient() error {
	a.clipboardMon.Stop()
	return a.wsClient.Disconnect()
}

// GetClientStatus 获取客户端状态
func (a *App) GetClientStatus() map[string]any {
	return map[string]any{
		"isConnected": a.wsClient.IsConnected(),
	}
}

// GetServerStatus 获取服务器状态
func (a *App) GetServerStatus() map[string]any {
	return map[string]any{
		"isRunning":   a.wsServer.IsRunning(),
		"clientCount": a.wsServer.GetClientCount(),
	}
}

// GetMode 获取当前模式
func (a *App) GetMode() string {
	return a.mode
}

// GetLocalIPs 获取本机 IP 地址列表
func (a *App) GetLocalIPs() ([]string, error) {
	return a.wsServer.GetLocalIPs()
}

// GetLogs 获取日志列表
func (a *App) GetLogs() []LogEntry {
	a.logsMu.RLock()
	defer a.logsMu.RUnlock()

	// 返回副本
	logs := make([]LogEntry, len(a.logs))
	copy(logs, a.logs)
	return logs
}

// ClearLogs 清空日志
func (a *App) ClearLogs() {
	a.logsMu.Lock()
	defer a.logsMu.Unlock()
	a.logs = make([]LogEntry, 0)
	runtime.EventsEmit(a.ctx, "logs:updated", a.logs)
}

// storeContent 暂存剪贴板内容，返回内容 ID（不可暂存时返回 0）
func (a *App) storeContent(dataType string, content []byte) int64 {
	if dataType != "text" && dataType != "image" {
		return 0
	}
	if dataType == "image" && len(content) > maxStoredImageBytes {
		return 0
	}

	a.contentsMu.Lock()
	defer a.contentsMu.Unlock()

	a.contentNext++
	id := a.contentNext
	cp := make([]byte, len(content))
	copy(cp, content)
	a.contents[id] = storedContent{dataType: dataType, content: cp}
	a.contentsOrder = append(a.contentsOrder, id)
	for len(a.contentsOrder) > maxStoredContents {
		delete(a.contents, a.contentsOrder[0])
		a.contentsOrder = a.contentsOrder[1:]
	}
	return id
}

// GetLogContent 按 ID 读取暂存的剪贴板内容（前端渲染缩略图/复制）
func (a *App) GetLogContent(id int64) map[string]any {
	a.contentsMu.RLock()
	defer a.contentsMu.RUnlock()

	sc, ok := a.contents[id]
	if !ok {
		return map[string]any{"found": false}
	}
	if sc.dataType == "text" {
		return map[string]any{"found": true, "type": "text", "text": string(sc.content)}
	}
	return map[string]any{
		"found":       true,
		"type":        "image",
		"imageBase64": base64.StdEncoding.EncodeToString(sc.content),
		"size":        len(sc.content),
	}
}

// CopyLogContent 把暂存的剪贴板内容写回系统剪贴板（一键复制）
func (a *App) CopyLogContent(id int64) error {
	a.contentsMu.RLock()
	sc, ok := a.contents[id]
	a.contentsMu.RUnlock()
	if !ok {
		return fmt.Errorf("内容已过期，不在暂存队列中")
	}

	mime := "text/plain"
	if sc.dataType == "image" {
		mime = "image/png"
	}
	if err := a.clipboardMon.SetClipboard(clipboard.ClipboardData{
		Type:     sc.dataType,
		Content:  sc.content,
		MimeType: mime,
	}); err != nil {
		return fmt.Errorf("写入剪贴板失败: %v", err)
	}
	return nil
}

// onLogFull 日志回调（websocket 包 LogCallback 使用）
func (a *App) onLogFull(e ws.LogEvent) {
	a.logsMu.Lock()
	defer a.logsMu.Unlock()

	category := e.Category
	if category == "" {
		category = "system"
	}

	// 同一条内容在一次同步链路里会被多层各记一行（收到 / 已写入 / 广播），
	// 按内容签名去重后只保留最先到达的那一行
	if category == "clipboard" && e.DataType != "" && e.Content != nil {
		if !a.clip.noteLogged(clipSig(e.DataType, e.Content)) {
			return
		}
	}

	var contentID int64
	if e.DataType != "" && e.Content != nil {
		contentID = a.storeContent(e.DataType, e.Content)
	}

	entry := LogEntry{
		Level:     e.Level,
		Message:   e.Message,
		Preview:   e.Preview,
		Category:  category,
		ContentID: contentID,
		Timestamp: time.Now().UnixMilli(),
	}

	a.logs = append([]LogEntry{entry}, a.logs...)
	if len(a.logs) > a.maxLogs {
		a.logs = a.logs[:a.maxLogs]
	}

	// 发送事件到前端
	runtime.EventsEmit(a.ctx, "logs:updated", a.logs)
}

// onLog 日志回调（无预览，分类 system）
func (a *App) onLog(level, message string) {
	a.onLogFull(ws.LogEvent{Level: level, Message: message, Category: "system"})
}

// onLogClip 日志回调（剪贴板相关，携带内容预览与原始数据，分类 clipboard）
func (a *App) onLogClip(level, message, dataType string, content []byte) {
	a.onLogFull(ws.LogEvent{
		Level:    level,
		Message:  message,
		Preview:  ws.BuildPreview(dataType, content),
		Category: "clipboard",
		DataType: dataType,
		Content:  content,
	})
}

// onClipboardChange 剪贴板变化回调（服务器模式，V1.1 二进制协议）
func (a *App) onClipboardChange(data clipboard.ClipboardData) {
	// 远程内容写回系统剪贴板会再次触发本地监听。这不是用户的新复制：
	// 若照常记日志并广播，同一条内容会在两端来回重复出现。
	if a.clip.isSelfEcho(clipSig(data.Type, data.Content)) {
		return
	}

	switch data.Type {
	case "text":
		a.onLogClip("INFO", fmt.Sprintf("检测到剪贴板文本变化: %d 字符", len(data.Content)), data.Type, data.Content)
	case "image":
		sizeMB := float64(len(data.Content)) / 1024 / 1024
		a.onLogClip("INFO", fmt.Sprintf("检测到剪贴板图片变化: %.2f MB", sizeMB), data.Type, data.Content)
	default:
		return
	}

	// 广播给所有客户端（V1.1 二进制协议）
	err := a.wsServer.BroadcastClipboardBinary(data.Type, data.Content)
	if err != nil {
		a.onLog("ERROR", fmt.Sprintf("广播剪贴板数据失败: %v", err))
	}
}

// onClipboardChangeClient 剪贴板变化回调（客户端模式，V1.1 二进制协议）
func (a *App) onClipboardChangeClient(data clipboard.ClipboardData) {
	if a.clip.isSelfEcho(clipSig(data.Type, data.Content)) {
		return
	}

	switch data.Type {
	case "text":
		a.onLogClip("INFO", fmt.Sprintf("检测到剪贴板文本变化: %d 字符", len(data.Content)), data.Type, data.Content)
	case "image":
		sizeMB := float64(len(data.Content)) / 1024 / 1024
		a.onLogClip("INFO", fmt.Sprintf("检测到剪贴板图片变化: %.2f MB", sizeMB), data.Type, data.Content)
	default:
		return
	}

	// 发送给服务器（V1.1 二进制协议）
	err := a.wsClient.SendClipboardBinary(data.Type, data.Content)
	if err != nil {
		a.onLog("ERROR", fmt.Sprintf("发送剪贴板数据失败: %v", err))
	}
}

// onClipboardReceivedBinary 接收到远程剪贴板数据回调（V1.1 二进制协议）
// dataType: "text" 或 "image"
// content: 原始二进制数据
func (a *App) onClipboardReceivedBinary(dataType string, content []byte) {
	// 将接收到的数据写入本地剪贴板
	data := clipboard.ClipboardData{
		Type:    dataType,
		Content: content,
	}

	switch dataType {
	case "text":
		data.MimeType = "text/plain"
	case "image":
		data.MimeType = "image/png"
	}

	// 先登记回声源：写回系统剪贴板必然触发本地监听，那份上报要当作同一条内容吞掉
	a.clip.markWritten(clipSig(dataType, content))

	err := a.clipboardMon.SetClipboard(data)
	if err != nil {
		a.onLog("ERROR", fmt.Sprintf("写入剪贴板失败: %v", err))
		return
	}

	switch dataType {
	case "text":
		a.onLogClip("SUCCESS", fmt.Sprintf("已接收并写入文本数据: %d 字符", len(content)), dataType, content)
	case "image":
		sizeMB := float64(len(content)) / 1024 / 1024
		a.onLogClip("SUCCESS", fmt.Sprintf("已接收并写入图片数据: %.2f MB", sizeMB), dataType, content)
	}
}
