package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// logStoreFileName 本地日志记录文件名
const logStoreFileName = "logs.json"

// logStoreDirName 本地数据目录名（位于用户配置目录下，不污染系统全局配置）
const logStoreDirName = "NextPaste"

// resolveLogStorePath 返回本地记录文件的绝对路径（必要时创建目录）
func resolveLogStorePath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, logStoreDirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, logStoreFileName), nil
}

// loadLogsFromDisk 从本地文件读取记录；文件不存在时返回空列表且不报错
func loadLogsFromDisk(path string) ([]LogEntry, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make([]LogEntry, 0), nil
		}
		return nil, err
	}
	if len(raw) == 0 {
		return make([]LogEntry, 0), nil
	}
	var logs []LogEntry
	if err := json.Unmarshal(raw, &logs); err != nil {
		// 文件损坏时退回空列表，避免因历史数据无法解析而无法启动
		return make([]LogEntry, 0), err
	}
	return logs, nil
}

// saveLogsToDisk 原子写入本地记录文件（先写临时文件再重命名）
func saveLogsToDisk(path string, logs []LogEntry) error {
	raw, err := json.Marshal(logs)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
