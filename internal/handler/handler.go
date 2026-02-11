package handler

import (
	"encoding/json"
	"io"
	"os"
)

// LoadConfig loads the config for the given environment (EnvProduccion or EnvDesarrollo).
// The JSON file must have a parent key "Produccion" or "Desarrollo" respectively.
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

// SaveConfig saves cfg to the config file for the given environment,
// writing only the corresponding parent key (Produccion or Desarrollo).
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
