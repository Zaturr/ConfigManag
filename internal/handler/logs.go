package handler

import (
	"os"
	"path/filepath"

	"github.com/SOLUCIONESSYCOM/scribe"
	"github.com/google/uuid"
)

var (
	apiLogger           *scribe.Scribe
	auditLogger         *scribe.Scribe
	initialGlobalFields map[string]interface{} // copia para añadir Usuario/Dispositivo al iniciar sesión
)

// Logs inicializa los loggers solo para el entorno seleccionado:
// - produccion → archivos en carpeta PROD
// - desarrollo → archivos en carpeta CERT
// Cualquier otro valor no crea loggers (apiLogger y auditLogger quedan nil).
func Logs(env string) error {
	if env != EnvProduccion && env != EnvDesarrollo {
		apiLogger = nil
		auditLogger = nil
		return nil
	}
	basePath := GetLogsBasePath(env)

	appGlobalFields := map[string]interface{}{
		"service_name":    "Soles_logger",
		"service_version": "1.0.0",
		"service_id":      uuid.New().String(),
		"env":             env,
		"path":            basePath,
	}
	initialGlobalFields = appGlobalFields
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
		Console:           false,
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

func SetSesionUsuario(usuario, dispositivo string) {
	if initialGlobalFields == nil {
		return
	}
	merged := make(map[string]interface{}, len(initialGlobalFields)+2)
	for k, v := range initialGlobalFields {
		merged[k] = v
	}
	merged["Usuario"] = usuario
	merged["Dispositivo"] = dispositivo
	scribe.SetGlobalFields(merged)
}

func LogSolestPOST(banco, url, body string) {
	if apiLogger == nil {
		return
	}
	apiLogger.Info().
		Str("banco", banco).
		Str("url", url).
		Str("body", body).
		Msg("POST solest config: body enviado al banco")
}

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

func LogHealthCheck(banco, url, resultado string, statusCode int, status, responseBody string) {
	if apiLogger == nil {
		return
	}
	response := responseBody
	if response == "" {
		response = resultado
	}
	ev := apiLogger.Info().
		Str("banco", banco).
		Str("url", url).
		Int("status_code", statusCode).
		Str("response", response)
	if status != "" {
		ev = ev.Str("status", status)
	}
	ev.Msg("Health check")
}

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

func GetUsuarioDispositivo() (usuario, dispositivo string) {
	usuario = os.Getenv("USERNAME")
	if usuario == "" {
		usuario = os.Getenv("USER")
	}
	if usuario == "" {
		usuario = "desconocido"
	}
	dispositivo, _ = os.Hostname()
	if dispositivo == "" {
		dispositivo = "desconocido"
	}
	return usuario, dispositivo
}

func LogSesionInicio(usuario, dispositivo, horaConexion string) {
	if auditLogger == nil {
		return
	}
	auditLogger.Info().
		Str("accion", "sesion_inicio").
		Str("usuario", usuario).
		Str("dispositivo", dispositivo).
		Str("hora_conexion", horaConexion).
		Msg("Sesión Solest Manager iniciada")
}

func LogSesionCierre(usuario, dispositivo, horaCierre string) {
	if auditLogger == nil {
		return
	}
	auditLogger.Info().
		Str("accion", "sesion_cierre").
		Str("usuario", usuario).
		Str("dispositivo", dispositivo).
		Str("hora_cierre", horaCierre).
		Msg("Sesión Solest Manager cerrada")
}
