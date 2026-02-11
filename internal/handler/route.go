package handler

import (
	"os"
	"path/filepath"
)

const (
	PathProduccion = "C:\\Users\\bdsyc\\OneDrive\\Escritorio\\PROD"
	PathDesarrollo = "C:\\Users\\bdsyc\\OneDrive\\Escritorio\\CERT"
	EnvProduccion  = "produccion"
	EnvDesarrollo  = "desarrollo"
)

func GetConfigPath(env string) (string, error) {
	if env == "" {
		env = EnvProduccion
	}
	var dir string
	switch env {
	case EnvDesarrollo:
		dir = PathDesarrollo
	case EnvProduccion:
		dir = PathProduccion
	default:
		dir = PathProduccion
	}
	if env == EnvProduccion {
		if override := os.Getenv("CONFIG_PATH_PROD"); override != "" {
			dir = override
		}
	} else if override := os.Getenv("CONFIG_PATH_CERT"); override != "" {
		dir = override
	}
	configPath := filepath.Join(filepath.Clean(dir), "config.json")
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
