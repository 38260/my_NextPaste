package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func sampleLogs(n int) []LogEntry {
	logs := make([]LogEntry, 0, n)
	for i := 0; i < n; i++ {
		logs = append(logs, LogEntry{
			Level:     "INFO",
			Message:   "msg",
			Preview:   "preview",
			Category:  "clipboard",
			Timestamp: int64(n - i),
		})
	}
	return logs
}

func TestSaveAndLoadLogsRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "logs.json")
	want := sampleLogs(3)
	if err := saveLogsToDisk(path, want); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	got, err := loadLogsFromDisk(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("entry %d = %+v, want %+v", i, got[i], want[i])
		}
	}
	// 不应残留临时文件
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Fatalf("temp file should be removed, err = %v", err)
	}
}

func TestLoadLogsMissingFileIsEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "not-exist.json")
	got, err := loadLogsFromDisk(path)
	if err != nil {
		t.Fatalf("missing file should not error, got %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("len = %d, want 0", len(got))
	}
}

func TestLoadLogsCorruptedFallsBackToEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "logs.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	got, err := loadLogsFromDisk(path)
	if err == nil {
		t.Fatalf("corrupted file should return an error")
	}
	if len(got) != 0 {
		t.Fatalf("len = %d, want 0", len(got))
	}
}

func TestApplyMaxLogsTrimsOldest(t *testing.T) {
	// logs 按时间倒序存放（最新在前），超量时从尾部删除最早记录
	a := &App{logs: sampleLogs(60), maxLogs: 60}
	a.applyMaxLogs(50)

	if len(a.logs) != 50 {
		t.Fatalf("len = %d, want 50", len(a.logs))
	}
	if a.logs[0].Timestamp != 60 || a.logs[49].Timestamp != 11 {
		t.Fatalf("kept window = [%d..%d], want [60..11]", a.logs[0].Timestamp, a.logs[49].Timestamp)
	}
	if a.maxLogs != 50 {
		t.Fatalf("maxLogs = %d, want 50", a.maxLogs)
	}
}

func TestApplyMaxLogsKeepsAllWhenUnderLimit(t *testing.T) {
	a := &App{logs: sampleLogs(3), maxLogs: 3}
	a.applyMaxLogs(50)
	if len(a.logs) != 3 {
		t.Fatalf("len = %d, want 3", len(a.logs))
	}
	if a.maxLogs != 50 {
		t.Fatalf("maxLogs = %d, want 50", a.maxLogs)
	}
}

func TestFlushLogsWritesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "logs.json")
	a := &App{logs: sampleLogs(2), maxLogs: 10, logStorePath: path}
	a.flushLogs()

	got, err := loadLogsFromDisk(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
}

func TestFlushLogsSkipsWhenPathEmpty(t *testing.T) {
	a := &App{logs: sampleLogs(1), maxLogs: 10}
	a.flushLogs() // 路径未就绪时不应 panic
}

func TestMarkLogsDirtyWritesWithinWindow(t *testing.T) {
	path := filepath.Join(t.TempDir(), "logs.json")
	a := &App{logs: sampleLogs(1), maxLogs: 10, logStorePath: path}
	a.markLogsDirty()

	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("记录未在合并窗口内落盘")
}
