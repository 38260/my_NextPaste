package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	ws "server/internal/websocket"
)

func recordTestApp(t *testing.T) *App {
	t.Helper()
	dir := t.TempDir()
	t.Chdir(dir)
	t.Setenv("APPDATA", filepath.Join(dir, "old-config"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "old-config"))
	if err := os.MkdirAll(filepath.Join(dir, "server"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "server", "wails.json"), []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	a := NewApp()
	a.initLogStore()
	stopRecordTimer(t, a)
	return a
}

func stopRecordTimer(t *testing.T, a *App) {
	t.Helper()
	t.Cleanup(func() {
		a.saveMu.Lock()
		if a.saveTimer != nil {
			a.saveTimer.Stop()
		}
		a.saveMu.Unlock()
		a.writeMu.Lock()
		a.writeMu.Unlock()
	})
}

func addRecord(a *App, dataType string, content []byte) int64 {
	a.onLogFull(ws.LogEvent{
		Level: "INFO", Category: "clipboard", DataType: dataType,
		Content: content, Preview: ws.BuildPreview(dataType, content),
	})
	return a.logs[0].ContentID
}

func TestRecordsRestoreFullTextImagesAndIDs(t *testing.T) {
	a := recordTestApp(t)
	text := strings.Repeat("中文全文\n", 100)
	textID := addRecord(a, "text", []byte(text))
	image, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+j4ioAAAAASUVORK5CYII=")
	if err != nil {
		t.Fatal(err)
	}
	imageID := addRecord(a, "image", image)
	for i := 0; i < 25; i++ {
		addRecord(a, "text", []byte(fmt.Sprintf("entry-%d", i)))
	}
	if err := a.saveLogStore(); err != nil {
		t.Fatal(err)
	}
	if len(a.contents) != 27 || a.contents[imageID].content != nil {
		t.Fatal("应保留全部内容元数据，已保存的图片不应常驻内存")
	}
	b := NewApp()
	b.initLogStore()
	stopRecordTimer(t, b)
	if got := b.GetLogContent(textID); got["text"] != text {
		t.Fatalf("全文未恢复: %+v", got)
	}
	got := b.GetLogContent(imageID)
	if got["imageBase64"] != base64.StdEncoding.EncodeToString(image) || got["size"] != len(image) {
		t.Fatalf("图片未恢复: %+v", got)
	}
	newID := addRecord(b, "text", []byte("new-after-restart"))
	if newID <= 27 || b.GetLogContent(textID)["text"] != text {
		t.Fatal("重启后 ID 被复用或旧内容被覆盖")
	}
}

func TestLegacyRecordsMigrateWithoutLosingHistory(t *testing.T) {
	a := recordTestApp(t)
	logs := sampleLogs(700)
	logs[0].ContentID = 900
	if err := saveLogsToDisk(a.logStorePath, logs); err != nil {
		t.Fatal(err)
	}
	b := NewApp()
	b.initLogStore()
	if len(b.logs) != 700 || b.logs[0].ContentID != 0 {
		t.Fatal("旧记录应全部加载，缺失原文的 ID 应清除")
	}
	stopRecordTimer(t, b)
	if id := addRecord(b, "text", []byte("new-content")); id <= 900 {
		t.Fatal("新 ID 不应复用旧格式中的 ID")
	}
	if err := b.saveLogStore(); err != nil {
		t.Fatal(err)
	}
	store, err := loadLogStore(b.logStorePath)
	if err != nil || len(store.Logs) != 700 || store.NextContentID != 901 {
		t.Fatalf("迁移失败: len=%d, next=%d, err=%v", len(store.Logs), store.NextContentID, err)
	}
}

func TestCorruptStoreIsNotOverwritten(t *testing.T) {
	a := recordTestApp(t)
	raw := []byte("{corrupt-records")
	if err := os.WriteFile(a.logStorePath, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	b := NewApp()
	b.initLogStore()
	if err := b.saveLogStore(); err == nil {
		t.Fatal("损坏的记录应报错并保留原文件")
	}
	got, err := os.ReadFile(a.logStorePath)
	if err != nil || !bytes.Equal(got, raw) {
		t.Fatal("损坏的原始文件被覆盖")
	}
}

func TestRecordLimitRestoresBeforeStartupTrimming(t *testing.T) {
	a := recordTestApp(t)
	if err := a.applyMaxLogs(700); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 710; i++ {
		addRecord(a, "text", []byte(fmt.Sprintf("content-%d", i)))
	}
	if err := a.saveLogStore(); err != nil {
		t.Fatal(err)
	}
	b := NewApp()
	b.initLogStore()
	if len(b.logs) != 700 || b.maxLogs != 700 || b.GetSettings()["maxLogs"] != 700 {
		t.Fatalf("启动前未恢复数量设置: logs=%d, max=%d", len(b.logs), b.maxLogs)
	}
	if b.GetLogContent(b.logs[699].ContentID)["text"] != "content-10" {
		t.Fatal("未按顺序保留最新 700 条")
	}
	if err := b.applyMaxLogs(50); err != nil {
		t.Fatal(err)
	}
	if len(b.contents) != 50 || b.GetLogContent(1)["found"] != false {
		t.Fatal("缩小数量后未同步淘汰内容")
	}
	files, err := os.ReadDir(filepath.Dir(contentFilePath(b.logStorePath, 1)))
	if err != nil || len(files) != 50 {
		t.Fatalf("内容文件未清理: len=%d, err=%v", len(files), err)
	}
	c := NewApp()
	c.initLogStore()
	if len(c.logs) != 50 || c.maxLogs != 50 {
		t.Fatal("数量缩小后未立即落盘")
	}
	if err := c.ClearLogs(); err != nil {
		t.Fatal(err)
	}
	d := NewApp()
	d.initLogStore()
	files, err = os.ReadDir(filepath.Dir(contentFilePath(d.logStorePath, 1)))
	if err != nil || len(d.logs) != 0 || len(d.contents) != 0 || len(files) != 0 {
		t.Fatal("清空后重启仍有记录或原文文件")
	}
	if id := addRecord(d, "text", []byte("after-clear")); id <= 710 {
		t.Fatal("清空并重启后复用了旧 ID")
	}
	stopRecordTimer(t, b)
	stopRecordTimer(t, c)
	stopRecordTimer(t, d)
}

func TestProjectStorageAndMigrationPreserveOldFiles(t *testing.T) {
	a := recordTestApp(t)
	cwd, _ := os.Getwd()
	if a.logStorePath != filepath.Join(cwd, "data", "logs.json") {
		t.Fatalf("应保存在项目 data 目录: %s", a.logStorePath)
	}
	if got := recordDataDir(filepath.Join(cwd, "server", "build", "bin"), t.TempDir()); got != cwd {
		t.Fatal("从其他工作目录启动时未找到项目根目录")
	}
	base, _ := os.UserConfigDir()
	oldPath := filepath.Join(base, "NextPaste", "logs.json")
	if err := os.MkdirAll(filepath.Dir(contentFilePath(oldPath, 7)), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := writeLocalFile(contentFilePath(oldPath, 7), []byte("migrated-text")); err != nil {
		t.Fatal(err)
	}
	store := logStore{Logs: []LogEntry{{ContentID: 7}}, MaxLogs: 800, NextContentID: 7, Contents: map[int64]string{7: "text"}}
	raw, err := json.Marshal(store)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeLocalFile(oldPath, raw); err != nil {
		t.Fatal(err)
	}
	b := NewApp()
	b.initLogStore()
	if b.maxLogs != 800 || b.GetLogContent(7)["text"] != "migrated-text" {
		t.Fatal("旧路径数据迁移失败")
	}
	got, err := os.ReadFile(oldPath)
	if err != nil || !bytes.Equal(got, raw) {
		t.Fatal("迁移修改了旧位置原文件")
	}
	if err := b.ClearLogs(); err != nil {
		t.Fatal(err)
	}
	c := NewApp()
	c.initLogStore()
	if len(c.logs) != 0 {
		t.Fatal("清空后不应再次迁移旧记录")
	}
}

func TestWriteFailurePreservesManifestAndContent(t *testing.T) {
	a := recordTestApp(t)
	id := addRecord(a, "text", []byte("saved-original"))
	if err := a.saveLogStore(); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(a.logStorePath)
	if err != nil {
		t.Fatal(err)
	}
	newID := addRecord(a, "text", []byte("pending-content"))
	if err := os.Mkdir(contentFilePath(a.logStorePath, newID), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := a.saveLogStore(); err == nil {
		t.Fatal("内容写入失败应向调用方返回错误")
	}
	got, err := os.ReadFile(a.logStorePath)
	if err != nil || !bytes.Equal(got, raw) || a.GetLogContent(id)["text"] != "saved-original" {
		t.Fatal("写入失败破坏了已经保存的记录")
	}
	if a.GetLogContent(newID)["text"] != "pending-content" {
		t.Fatal("写入失败时待保存内容被丢弃")
	}
}

func TestRetentionIgnoresUnrelatedFiles(t *testing.T) {
	a := recordTestApp(t)
	id := addRecord(a, "text", []byte("clear-me"))
	if err := a.saveLogStore(); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Dir(contentFilePath(a.logStorePath, id))
	path := filepath.Join(dir, "user-notes.txt")
	if err := os.WriteFile(path, []byte("preserve"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := a.ClearLogs(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal("清理删除了不属于记录的文件")
	}
}

func TestFailedRetentionKeepsMemoryAndDiskRecords(t *testing.T) {
	a := recordTestApp(t)
	for i := 0; i < 60; i++ {
		addRecord(a, "text", []byte(fmt.Sprintf("preserved-%d", i)))
	}
	if err := a.saveLogStore(); err != nil {
		t.Fatal(err)
	}
	path := a.logStorePath
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	a.logStorePath = filepath.Join(t.TempDir(), "missing-dir", "logs.json")
	if err := a.ClearLogs(); err == nil {
		t.Fatal("清空写入失败应报错")
	}
	if err := a.applyMaxLogs(50); err == nil {
		t.Fatal("数量修改写入失败应报错")
	}
	a.logStorePath = path
	got, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(got, raw) || len(a.logs) != 60 || len(a.contents) != 60 || a.maxLogs != 500 {
		t.Fatal("写入失败丢弃了原内存或磁盘记录")
	}
	if a.GetLogContent(a.logs[59].ContentID)["text"] != "preserved-0" {
		t.Fatal("写入失败后原文不可读")
	}
}
