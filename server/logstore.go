package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// logStoreFileName 本地日志记录文件名
const logStoreFileName = "logs.json"

// logStoreDirName 本地数据目录名（位于用户配置目录下，不污染系统全局配置）
const logStoreDirName = "NextPaste"

type logStore struct {
	Logs          []LogEntry       `json:"logs"`
	MaxLogs       int              `json:"maxLogs"`
	NextContentID int64            `json:"nextContentId"`
	Contents      map[int64]string `json:"contents"`
}

// resolveLogStorePath 返回本地记录文件的绝对路径（必要时创建目录）
func resolveLogStorePath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, logStoreDirName)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return filepath.Join(dir, logStoreFileName), nil
}

// loadLogsFromDisk 从本地文件读取记录；文件不存在时返回空列表且不报错
func loadLogsFromDisk(path string) ([]LogEntry, error) {
	store, err := loadLogStore(path)
	return store.Logs, err
}

// loadLogStore 兼容旧版日志数组；旧版不按默认 500 条提前裁剪。
func loadLogStore(path string) (logStore, error) {
	store := logStore{Logs: make([]LogEntry, 0), MaxLogs: defaultMaxLogs, Contents: make(map[int64]string)}
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return store, nil
		}
		return store, err
	}
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return store, nil
	}
	if raw[0] == '[' {
		store.MaxLogs = 0
		err = json.Unmarshal(raw, &store.Logs)
	} else if raw[0] == '{' {
		err = json.Unmarshal(raw, &store)
	} else {
		err = fmt.Errorf("不支持的记录文件格式")
	}
	if err != nil {
		return logStore{Logs: make([]LogEntry, 0)}, err
	}
	for id, dataType := range store.Contents {
		if id <= 0 || (dataType != "text" && dataType != "image") {
			return logStore{Logs: make([]LogEntry, 0)}, fmt.Errorf("无效的内容记录: %d", id)
		}
	}
	if store.MaxLogs <= 0 {
		store.MaxLogs = max(defaultMaxLogs, len(store.Logs))
	}
	store.MaxLogs = min(max(store.MaxLogs, minMaxLogs), maxMaxLogs)
	for _, entry := range store.Logs {
		store.NextContentID = max(store.NextContentID, entry.ContentID)
	}
	for id := range store.Contents {
		store.NextContentID = max(store.NextContentID, id)
	}
	return store, nil
}

// saveLogsToDisk 原子写入本地记录文件（先写临时文件再重命名）
func saveLogsToDisk(path string, logs []LogEntry) error {
	raw, err := json.Marshal(logs)
	if err != nil {
		return err
	}
	return writeLocalFile(path, raw)
}

// writeLocalFile 失败时保留旧文件，临时文件只在写完并关闭后替换目标。
func writeLocalFile(path string, data []byte) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".nextpaste-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func contentFilePath(path string, id int64) string {
	return filepath.Join(filepath.Dir(path), "contents", strconv.FormatInt(id, 10)+".bin")
}
