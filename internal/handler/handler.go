package handler

import (
	"encoding/json"
	"io"
	"os"
)

func LoadConfig(env string) (Config, error) {
	path, err := GetConfigPath(env)
	if err != nil {
		return nil, err
	}
	if !ConfigExists(path) {
		return make(Config), nil
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var file ConfigFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	if env == EnvDesarrollo {
		if file.Desarrollo == nil {
			return make(Config), nil
		}
		return file.Desarrollo, nil
	}
	if file.Produccion == nil {
		return make(Config), nil
	}
	return file.Produccion, nil
}

func SaveConfig(env string, cfg Config) error {
	path, err := GetConfigPath(env)
	if err != nil {
		return err
	}

	var file ConfigFile
	if env == EnvDesarrollo {
		file = ConfigFile{Desarrollo: cfg}
	} else {
		file = ConfigFile{Produccion: cfg}
	}
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
