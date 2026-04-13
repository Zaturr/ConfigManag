package handler

import (
	//"encoding/json"
	"errors"
	//"io"
	"os"
	"reflect"

	"github.com/charmbracelet/huh"
)

var ExpectedPassword = "123456"

//var ExpectedPassword = os.Getenv("CONFIG_PASSWORD")

// var (
//
//	credentialsConfigPath   string
//	credentialsConfigLoaded bool
//
// )
//
//	func loadCredentialsConfig(path string) (string, error) {
//		if path == "" {
//			path = "Credentials\\config.json"
//		}
//		f, err := os.Open(path)
//		if err != nil {
//			return "", err
//		}
//		defer f.Close()
//
//		data, err := io.ReadAll(f)
//		if err != nil {
//			return "", err
//		}
//		var cfg struct {
//			Password string `json:"password"`
//		}
//		if err := json.Unmarshal(data, &cfg); err != nil {
//			return "", err
//		}
//		return cfg.Password, nil
//	}
//
// func CredentialsConfigPath() string { return credentialsConfigPath }
// func CredentialsConfigLoaded() bool { return credentialsConfigLoaded }
//
//	func init() {
//		path := "Credentials\\config.json"
//		credentialsConfigPath = path
//		pwd, err := loadCredentialsConfig(path)
//		if err == nil && pwd != "" {
//			ExpectedPassword = pwd
//			credentialsConfigLoaded = true
//		}
//	}
func GetPassword() (string, error) {
	var password string
	expected := ExpectedPassword
	if expected == "" {
		expected = os.Getenv("CONFIG_PASSWORD")
	}
	input := huh.NewInput().
		Title("Introduce tu contraseña").
		Value(&password).
		EchoMode(huh.EchoModePassword).
		Validate(func(s string) error { return validate(s, expected) })

	setPasswordEchoChar(input, '*')

	err := huh.NewForm(huh.NewGroup(input)).Run()
	if err != nil {
		return "", err
	}

	ResetInactivity()
	return password, nil
}

func validate(s, expected string) error {
	if s == "" {
		return errors.New("la contraseña no puede estar vacía")
	}
	if expected != "" && s != expected {
		return errors.New("contraseña incorrecta")
	}
	return nil
}

func setPasswordEchoChar(input *huh.Input, char rune) {
	v := reflect.ValueOf(input).Elem()

	ti := v.FieldByName("textinput")
	if ti.IsValid() {
		ec := ti.FieldByName("EchoCharacter")
		if ec.IsValid() && ec.CanSet() {
			ec.Set(reflect.ValueOf(char))
		}
	}
}
