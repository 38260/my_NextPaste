package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
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

// 本地记录条数限制（可在设置中调整，超出后按时间顺序删除最早的记录）
const (
	defaultMaxLogs = 500
	minMaxLogs     = 50
	maxMaxLogs     = 10000
)

// logSaveDelay 记录落盘的合并窗口（避免高频日志频繁写盘）
const logSaveDelay = 800 * time.Millisecond

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

	contents    map[int64]storedContent
	contentsMu  sync.RWMutex
	contentNext int64

	clip *clipLedger

	// 记录本地持久化
	logStorePath string
	logStoreErr  error
	saveMu       sync.Mutex
	saveTimer    *time.Timer
	writeMu      sync.Mutex

	launchFlags LaunchFlags
	trayReady   atomic.Bool
	quitting    atomic.Bool

	// 托盘右键菜单当前展示的记录（弹出时重建，点击时按命令 ID 取回）
	trayMenuMu      sync.Mutex
	trayMenuRecords []trayRecord

	// 前端“显示内容”开关的同步副本：关闭时托盘菜单不展示剪贴板内容摘要
	contentRevealed atomic.Bool
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		wsServer:     ws.NewServer(),
		wsClient:     ws.NewWSClient("NextPaste Desktop", "Windows"),
		clipboardMon: clipboard.NewMonitor(),
		logs:         make([]LogEntry, 0),
		maxLogs:      defaultMaxLogs,
		contents:     make(map[int64]storedContent),
		clip:         newClipLedger(),
		mode:         "server", // 默认为服务器模式
		settings: Settings{
			AutoStartEnabled:       false,
			AutoStartActionEnabled: false,
			AutoStartMode:          AutoStartModeServer,
			CloseAction:            "tray",
			MaxLogs:                defaultMaxLogs,
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

// showSettings 显示主窗口并打开设置面板（托盘菜单“打开设置”入口）
func (a *App) showSettings() {
	if a.ctx == nil {
		return
	}
	a.ShowWindow()
	runtime.EventsEmit(a.ctx, "settings:open")
}

// showRecord 显示主窗口并定位到指定剪贴板记录（托盘菜单点击记录）
func (a *App) showRecord(id int64) {
	if a.ctx == nil {
		return
	}
	a.ShowWindow()
	runtime.EventsEmit(a.ctx, "record:focus", id)
}

// onSecondInstanceLaunch 单实例锁回调：重复启动时唤起已有窗口
func (a *App) onSecondInstanceLaunch() {
	if a.ctx == nil {
		return
	}
	a.ShowWindow()
}

// autostartEnabled 查询开机自启勾选态（读注册表，以系统为准）
func (a *App) autostartEnabled() bool {
	enabled, _, err := isAutostartEnabled()
	if err != nil {
		log.Printf("query autostart for tray menu failed: %v", err)
	}
	return err == nil && enabled
}

// toggleTrayAutostart 托盘菜单切换开机自启；开启时沿用当前工作模式，成功后通知前端同步
func (a *App) toggleTrayAutostart() {
	enabled, mode, err := isAutostartEnabled()
	if err != nil {
		a.onLog("ERROR", "托盘切换开机自启失败：读取注册表状态出错 "+err.Error())
		return
	}
	if !enabled {
		a.settingsMu.RLock()
		mode = a.settings.AutoStartMode
		a.settingsMu.RUnlock()
	}
	next := !enabled
	if err := setAutostart(next, mode); err != nil {
		a.onLog("ERROR", fmt.Sprintf("托盘%s开机自启失败: %v", map[bool]string{true: "开启", false: "关闭"}[next], err))
		return
	}
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "autostart:changed", next)
	}
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
	// 记录条数：缺省（<=0）沿用当前值，越界收敛到允许区间
	if s.MaxLogs <= 0 {
		a.settingsMu.RLock()
		s.MaxLogs = a.settings.MaxLogs
		a.settingsMu.RUnlock()
	}
	if s.MaxLogs < minMaxLogs {
		s.MaxLogs = minMaxLogs
	}
	if s.MaxLogs > maxMaxLogs {
		s.MaxLogs = maxMaxLogs
	}
	// 关闭偏好必须先生效；开机自启写入失败不应让窗口行为退回默认值
	a.settingsMu.Lock()
	a.settings.CloseAction = s.CloseAction
	a.settingsMu.Unlock()
	if err := a.applyMaxLogs(s.MaxLogs); err != nil {
		return err
	}
	a.settingsMu.Lock()
	a.settings.MaxLogs = s.MaxLogs
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
		"maxLogs":                a.settings.MaxLogs,
	}
}

// GetLogStorePath 返回本地记录文件路径（设置面板展示“数据存在哪里”）
func (a *App) GetLogStorePath() string {
	return a.logStorePath
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
	a.initLogStore()
	trayReady := a.startTray()

	// Settings live in localStorage on the frontend; backend stores the last applied copy.
	// Auto start/connect is triggered from the frontend (so it can reuse persisted
	// address/port/url and show UI errors).
	if a.launchFlags.Minimized {
		select {
		case <-trayReady:
		case <-time.After(5 * time.Second):
		}
		// 隐藏启动但托盘不可用时，先显示窗口以保留任务栏恢复入口。
		if !a.trayReady.Load() {
			runtime.WindowShow(a.ctx)
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
	// 退出前把待写记录落盘，避免丢最后一批日志
	a.flushLogs()
}

// ============================================
// 托盘右键菜单：最近剪贴板记录
// ============================================

// trayRecord 托盘右键菜单中的最近剪贴板记录条目
type trayRecord struct {
	ID      int64  `json:"id"`      // 关联的暂存内容 ID
	Label   string `json:"label"`   // 菜单标题（时间 + 摘要）
	Tooltip string `json:"tooltip"` // 菜单悬停提示（时间 + 完整预览）
}

// trayRecentCount 托盘右键菜单展示的最近剪贴板记录条数
const trayRecentCount = 3

// trayMenuLabelLimit 菜单标题中预览摘要的最大长度（rune）
const trayMenuLabelLimit = 32

// trayMaskedSummary 内容隐藏偏好开启时，托盘菜单中替代内容摘要的占位文案
const trayMaskedSummary = "内容已隐藏"

// SetContentRevealed 同步前端“显示内容”开关，决定托盘菜单是否展示内容摘要
func (a *App) SetContentRevealed(revealed bool) {
	a.contentRevealed.Store(revealed)
}

// recentClipRecords 返回最近 n 条剪贴板记录（最新在前），供托盘菜单展示
func (a *App) recentClipRecords(n int) []trayRecord {
	a.logsMu.RLock()
	defer a.logsMu.RUnlock()

	revealed := a.contentRevealed.Load()
	records := make([]trayRecord, 0, n)
	for _, entry := range a.logs {
		if entry.Category != "clipboard" || entry.ContentID == 0 {
			continue
		}
		title := time.UnixMilli(entry.Timestamp).Format("15:04:05")
		summary := trayMaskedSummary
		if revealed {
			summary = truncateRunes(entry.Preview, trayMenuLabelLimit)
		}
		records = append(records, trayRecord{
			ID:      entry.ContentID,
			Label:   title + " · " + summary,
			Tooltip: title + " · " + summary,
		})
		if len(records) >= n {
			break
		}
	}
	return records
}

// truncateRunes 按 rune 截断超长文本并补省略号
func truncateRunes(text string, limit int) string {
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	return string(runes[:limit]) + "…"
}

// ============================================
// 记录本地持久化
// ============================================

// initLogStore 解析本地记录文件路径并加载历史记录
func (a *App) initLogStore() {
	path, err := resolveLogStorePath()
	if err != nil {
		log.Printf("resolve log store path failed: %v", err)
		a.logStoreErr = err
		return
	}
	a.logStorePath = path
	if err := migrateLogStore(path); err != nil {
		a.logStoreErr = err
		log.Printf("migrate logs failed: %v", err)
		return
	}

	store, err := loadLogStore(path)
	if err != nil {
		log.Printf("load logs failed: %v", err)
		a.logStoreErr = err // 不自动覆盖无法读取的旧数据。
		return
	}

	a.logsMu.Lock()
	a.maxLogs = store.MaxLogs
	a.settings.MaxLogs = store.MaxLogs
	a.logs = store.Logs
	a.contentNext = store.NextContentID
	for i := range a.logs {
		id := a.logs[i].ContentID
		dataType, ok := store.Contents[id]
		if ok {
			_, err = os.Stat(contentFilePath(path, id))
		}
		if !ok || err != nil {
			a.logs[i].ContentID = 0 // 旧版只有预览，不能让旧 ID 指向新内容。
			continue
		}
		a.contents[id] = storedContent{dataType: dataType}
	}
	if len(a.logs) > a.maxLogs {
		a.logs = a.logs[:a.maxLogs]
	}
	a.pruneContents()
	snapshot := make([]LogEntry, len(a.logs))
	copy(snapshot, a.logs)
	a.logsMu.Unlock()

	if a.ctx != nil && len(snapshot) > 0 {
		runtime.EventsEmit(a.ctx, "logs:updated", snapshot)
	}
}

// markLogsDirty 合并窗口内的高频日志，最多每 logSaveDelay 落盘一次
func (a *App) markLogsDirty() {
	a.saveMu.Lock()
	defer a.saveMu.Unlock()
	if a.saveTimer != nil {
		return // 已有待写任务，本次合并进去
	}
	a.saveTimer = time.AfterFunc(logSaveDelay, func() {
		a.saveMu.Lock()
		a.saveTimer = nil
		a.saveMu.Unlock()
		a.flushLogs()
	})
}

// flushLogs 立即把当前记录写入本地文件（原子写）
func (a *App) flushLogs() {
	if err := a.saveLogStore(); err != nil {
		log.Printf("save logs failed: %v", err)
	}
}

func (a *App) saveLogStore() error {
	// 在取快照前串行化写入，避免较早的快照覆盖清空或缩小数量后的新状态。
	a.writeMu.Lock()
	defer a.writeMu.Unlock()
	a.logsMu.RLock()
	store, pending := a.recordSnapshot(a.logs, a.maxLogs)
	a.logsMu.RUnlock()
	if err := a.writeRecordSnapshot(store, pending); err != nil {
		return err
	}
	return a.finishLogSave(store, pending)
}

// recordSnapshot 调用方持有 logsMu，快照中的内容字节不再修改。
func (a *App) recordSnapshot(logs []LogEntry, maxLogs int) (logStore, map[int64][]byte) {
	a.contentsMu.RLock()
	defer a.contentsMu.RUnlock()
	store := logStore{
		Logs: append([]LogEntry{}, logs...), MaxLogs: maxLogs,
		NextContentID: a.contentNext, Contents: make(map[int64]string),
	}
	pending := make(map[int64][]byte)
	for _, entry := range store.Logs {
		if sc, ok := a.contents[entry.ContentID]; ok {
			store.Contents[entry.ContentID] = sc.dataType
			if sc.content != nil {
				pending[entry.ContentID] = sc.content
			}
		}
	}
	return store, pending
}

func (a *App) writeRecordSnapshot(store logStore, pending map[int64][]byte) error {
	if a.logStoreErr != nil {
		return fmt.Errorf("原记录文件无法读取，已保留原文件: %w", a.logStoreErr)
	}
	path := a.logStorePath
	if path == "" {
		return nil
	}
	if len(pending) > 0 {
		if err := os.MkdirAll(filepath.Join(filepath.Dir(path), "contents"), 0o700); err != nil {
			return err
		}
	}
	for id, content := range pending {
		if err := writeLocalFile(contentFilePath(path, id), content); err != nil {
			return fmt.Errorf("保存剪贴板内容失败: %w", err)
		}
	}
	raw, err := json.Marshal(store)
	if err != nil {
		return err
	}
	if err := writeLocalFile(path, raw); err != nil {
		return err
	}
	return nil
}

func (a *App) finishLogSave(store logStore, pending map[int64][]byte) error {
	if a.logStorePath == "" {
		return nil
	}
	// 已落盘内容按需读取，不让保留数量决定常驻图片内存。
	a.contentsMu.Lock()
	for id := range pending {
		if sc, ok := a.contents[id]; ok {
			sc.content = nil
			a.contents[id] = sc
		}
	}
	a.contentsMu.Unlock()
	return cleanupContentFiles(a.logStorePath, store)
}

// applyMaxLogs 更新记录上限并按序裁剪（超出部分删除最早的记录）
func (a *App) applyMaxLogs(n int) error {
	n = min(max(n, minMaxLogs), maxMaxLogs)
	return a.retainLogs(n)
}

// retainLogs 先保存候选记录，再更新内存；n=0 清空但保留数量设置。
func (a *App) retainLogs(n int) error {
	a.writeMu.Lock()
	defer a.writeMu.Unlock()
	a.logsMu.Lock()
	logs, maxLogs := a.logs, a.maxLogs
	if n == 0 {
		logs = nil
	} else {
		maxLogs = n
	}
	if len(logs) > n {
		logs = logs[:n]
	}
	store, pending := a.recordSnapshot(logs, maxLogs)
	if err := a.writeRecordSnapshot(store, pending); err != nil {
		a.logsMu.Unlock()
		return err
	}
	a.logs, a.maxLogs = logs, maxLogs
	a.pruneContents()
	a.logsMu.Unlock()

	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "logs:updated", store.Logs)
	}
	return a.finishLogSave(store, pending)
}

// pruneContents 调用方持有 logsMu，只保留仍被记录引用的内容。
func (a *App) pruneContents() {
	retained := make(map[int64]bool, len(a.logs))
	for _, entry := range a.logs {
		retained[entry.ContentID] = true
	}
	a.contentsMu.Lock()
	defer a.contentsMu.Unlock()
	for id := range a.contents {
		if !retained[id] {
			delete(a.contents, id)
		}
	}
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

// GetServerClients 获取已连接设备列表（服务器模式下“本机作为服务器”的设备明细）
func (a *App) GetServerClients() []map[string]any {
	return a.wsServer.GetClients()
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
func (a *App) ClearLogs() error {
	return a.retainLogs(0)
}

// storeContent 暂存剪贴板内容，返回内容 ID（不可暂存时返回 0）
func (a *App) storeContent(dataType string, content []byte) int64 {
	if dataType != "text" && dataType != "image" {
		return 0
	}

	a.contentsMu.Lock()
	defer a.contentsMu.Unlock()

	a.contentNext++
	id := a.contentNext
	cp := make([]byte, len(content))
	copy(cp, content)
	a.contents[id] = storedContent{dataType: dataType, content: cp}
	return id
}

func (a *App) readContent(id int64) (storedContent, error) {
	a.contentsMu.RLock()
	defer a.contentsMu.RUnlock()
	sc, ok := a.contents[id]
	if !ok {
		return storedContent{}, fmt.Errorf("内容已过期或旧版未保存原文")
	}
	if sc.content == nil {
		var err error
		sc.content, err = os.ReadFile(contentFilePath(a.logStorePath, id))
		if err != nil {
			return storedContent{}, fmt.Errorf("读取本地内容失败: %w", err)
		}
	}
	return sc, nil
}

// GetLogContent 按 ID 读取暂存的剪贴板内容（前端渲染缩略图/复制）
func (a *App) GetLogContent(id int64) map[string]any {
	sc, err := a.readContent(id)
	if err != nil {
		return map[string]any{"found": false, "error": err.Error()}
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
	sc, err := a.readContent(id)
	if err != nil {
		return err
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
		a.pruneContents()
	}

	// 发送事件到前端
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, "logs:updated", a.logs)
	}

	// 本地落盘（合并窗口内只写一次）
	a.markLogsDirty()
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
