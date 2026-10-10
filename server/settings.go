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
	CloseAction            string        `json:"closeAction"` // "minimize" | "close"
}

func getExecPath() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(path)
}
