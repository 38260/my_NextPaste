package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"testing"

	ws "server/internal/websocket"
)

func recordTestApp(t *testing.T) *App {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("APPDATA", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)
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
