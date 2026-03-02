package handler

import (
	"github.com/SOLUCIONESSYCOM/scribe"
	"github.com/google/uuid"
)

var (
	apiLogger   *scribe.Scribe
	auditLogger *scribe.Scribe
)

func Logs() error {
	appGlobalFields := map[string]interface{}{
		"service_name":    "Soles_logger",
		"service_version": "1.0.0",
		"service_id":      uuid.New().String(),
	}
	scribe.SetGlobalFields(appGlobalFields)

	config := &scribe.ConfigLogger{
		FilePath:          "./logs/api",
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
		FilePath:          "./logs/audit",
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

	apiLogger.Info().Msg("API Logger initialized")
	auditLogger.Info().Str("service_name", "Soles_Audit_logger").Msg("Audit Logger initialized")

	return nil
}
