package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// logStoreFileName 本地日志记录文件名
const logStoreFileName = "logs.json"

type logStore struct {
	Logs          []LogEntry       `json:"logs"`
	MaxLogs       int              `json:"maxLogs"`
	NextContentID int64            `json:"nextContentId"`
	Contents      map[int64]string `json:"contents"`
}

// resolveLogStorePath 返回本地记录文件的绝对路径（必要时创建目录）
func resolveLogStorePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(recordDataDir(filepath.Dir(exe), cwd), "data")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return filepath.Join(dir, logStoreFileName), nil
}

// recordDataDir 正式版从程序位置寻找项目根目录，开发时回退到工作目录。
func recordDataDir(exeDir, cwd string) string {
	for _, start := range []string{exeDir, cwd} {
		for dir := start; ; dir = filepath.Dir(dir) {
			if info, err := os.Stat(filepath.Join(dir, "server", "wails.json")); err == nil && !info.IsDir() {
				return dir
			}
			if filepath.Dir(dir) == dir {
				break
			}
		}
	}
	return exeDir // 独立分发时数据与程序放在同一目录。
}

// migrateLogStore 只在新位置没有记录时复制旧数据，保留旧位置原文件。
func migrateLogStore(path string) error {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return err
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	oldPath := filepath.Join(base, "NextPaste", logStoreFileName)
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	store, err := loadLogStore(oldPath)
	if err != nil {
		return fmt.Errorf("读取旧位置记录失败，原文件已保留: %w", err)
	}
	for id := range store.Contents {
		data, err := os.ReadFile(contentFilePath(oldPath, id))
		if os.IsNotExist(err) {
			delete(store.Contents, id)
			continue
		}
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(contentFilePath(path, id)), 0o700); err != nil {
			return err
		}
		if err := writeLocalFile(contentFilePath(path, id), data); err != nil {
			return err
		}
	}
	raw, err := json.Marshal(store)
	if err != nil {
		return err
	}
	return writeLocalFile(path, raw)
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

// cleanupContentFiles 仅清理本应用生成的数字编号文件，且必须在记录列表保存成功后执行。
func cleanupContentFiles(path string, store logStore) error {
	dir := filepath.Dir(contentFilePath(path, 1))
	files, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, file := range files {
		if !file.Type().IsRegular() || filepath.Ext(file.Name()) != ".bin" {
			continue
		}
		id, err := strconv.ParseInt(strings.TrimSuffix(file.Name(), ".bin"), 10, 64)
		if err != nil || id <= 0 || id > store.NextContentID {
			continue
		}
		if _, retained := store.Contents[id]; !retained {
			if err := os.Remove(filepath.Join(dir, file.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}
