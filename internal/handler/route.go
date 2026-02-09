package handler

import (
	"os"
	"path/filepath"
	"strings"
)

const ConfigRelPath = "C:\\Users\\bdsyc\\OneDrive\\Escritorio\\config"

func resolveConfigPath(p string, cwd string) string {
	if p == "" {
		return ""
	}
	path := p
	if !filepath.IsAbs(p) {
		path = filepath.Join(cwd, p)
	}
	if !strings.HasSuffix(strings.ToLower(path), ".json") {
		path = filepath.Join(path, "config.json")
	}
	return path
}

func GetConfigPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	p := os.Getenv("CONFIG_PATH")
	if p == "" {
		p = ConfigRelPath
	}

	configPath := resolveConfigPath(p, cwd)
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}
	return configPath, nil
}

func ConfigExists(configPath string) bool {
	_, err := os.Stat(configPath)
	return err == nil
}
