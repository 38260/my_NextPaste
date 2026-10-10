package main

import (
	"os"
	"path/filepath"
)

type AutoStartMode string

const (
	AutoStartModeServer AutoStartMode = "server"
	AutoStartModeClient AutoStartMode = "client"
)

type Settings struct {
	AutoStartEnabled       bool          `json:"autoStartEnabled"`
	AutoStartActionEnabled bool          `json:"autoStartActionEnabled"`
	AutoStartMode          AutoStartMode `json:"autoStartMode"`
	CloseAction            string        `json:"closeAction"` // "tray" | "close"
	MaxLogs                int           `json:"maxLogs"`     // 本地保留的记录条数（超出按时间顺序删除最早的）
}

func getExecPath() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(path)
}
