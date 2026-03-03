package handler

import (
	"os"
	"path/filepath"

	"github.com/SOLUCIONESSYCOM/scribe"
	"github.com/google/uuid"
)

var (
	apiLogger   *scribe.Scribe
	auditLogger *scribe.Scribe
)

func Logs() error {
	env := os.Getenv("CONFIG_ENV")
	if env == "" {
		env = EnvProduccion
	}
	basePath := GetLogsBasePath(env)

	appGlobalFields := map[string]interface{}{
		"service_name":    "Soles_logger",
		"service_version": "1.0.0",
		"service_id":      uuid.New().String(),
		"env":             env,
	}
	scribe.SetGlobalFields(appGlobalFields)

	config := &scribe.ConfigLogger{
		FilePath:          filepath.Join(basePath, "api"),
		MinLevel:          "debug",
		RotationMaxSizeMB: 10,
		MaxBackups:        5,
		MaxAgeDay:         30,
		Compress:          false,
		Console:           false,
		BeutifyConsoleLog: false,
		File:              true,
	}

	var err error
	apiLogger, err = scribe.New(config, nil, nil)
	if err != nil {
		return err
	}

	auditConfig := &scribe.ConfigLogger{
		FilePath:          filepath.Join(basePath, "audit"),
		MinLevel:          "info",
		RotationMaxSizeMB: 10,
		MaxBackups:        5,
		MaxAgeDay:         30,
		Compress:          false,
		Console:           true,
		BeutifyConsoleLog: false,
		File:              true,
	}

	auditLogger, err = scribe.New(auditConfig, nil, nil)
	if err != nil {
		return err
	}

	scribe.SetDefaultLogger(apiLogger)

	apiLogger.Info().Str("env", env).Str("path", basePath).Msg("API Logger initialized")
	auditLogger.Info().Str("service_name", "Soles_Audit_logger").Str("env", env).Str("path", basePath).Msg("Audit Logger initialized")

	return nil
}

// LogTrace escribe en el log normal (api) una traza con mensaje y detalles opcionales.
// Úsalo para flujo completo: menús, pasos, decisiones, errores recuperables, etc.
func LogTrace(msg string, detalles map[string]interface{}) {
	if apiLogger == nil {
		return
	}
	ev := apiLogger.Info()
	for k, v := range detalles {
		ev = ev.Interface(k, v)
	}
	ev.Msg(msg)
}

// LogAccion registra en el log de auditoría el final de una acción (entorno, acción y detalles).
// Solo para eventos que deban quedar en auditoría (cambios de config, acciones críticas).
func LogAccion(accion, entorno string, detalles map[string]interface{}) {
	if auditLogger == nil {
		return
	}
	ev := auditLogger.Info().Str("accion", accion).Str("entorno", entorno)
	for k, v := range detalles {
		ev = ev.Interface(k, v)
	}
	ev.Msg("Acción realizada")
}
