package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"v2/internal/handler"

	"github.com/gin-gonic/gin"
)

// Puerto del API del banco para solest config (se usa junto con la IP del config).
const solestPort = "8082"

func SolestConfig(cfg handler.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		nombreBanco := ctx.Query("banco")
		if nombreBanco == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "falta query: ?banco=NombreDelBanco"})
			return
		}

		// Si viene env, usar config recién guardada en disco (p. ej. llamada desde main)
		cfgUsar := cfg
		if envParam := ctx.Query("env"); envParam != "" {
			carga, err := handler.LoadConfig(envParam)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error cargando config: " + err.Error()})
				return
			}
			cfgUsar = carga
		}

		banco, ok := cfgUsar[nombreBanco]
		if !ok {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "banco no encontrado: " + nombreBanco})
			return
		}

		// Config: solo IP (ej. 192.168.100.230) y endpoint como path (ej. /simf/api/v1/config/solest).
		// Se arma la URL aquí: http:// + IP + :8082 + endpoint
		ip := strings.TrimSpace(banco.IP)
		endpointPath := strings.TrimSpace(banco.Endpoint)
		if ip == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("banco %q no tiene IP configurada en el archivo de configuración", nombreBanco),
			})
			return
		}
		if endpointPath == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("banco %q no tiene endpoint configurado (path, ej. /simf/api/v1/config/solest)", nombreBanco),
			})
			return
		}
		if !strings.HasPrefix(endpointPath, "/") {
			endpointPath = "/" + endpointPath
		}
		urlBanco := "http://" + ip + ":" + solestPort + endpointPath

		BodyEnvio, err := json.Marshal(banco.Envio)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error armando JSON de envio: " + err.Error()})
			return

		}

		// Log solo a archivo (no consola) vía apiLogger
		handler.LogSolestPOST(nombreBanco, urlBanco, string(BodyEnvio))

		req, err := http.NewRequest(http.MethodPost, urlBanco, bytes.NewReader(BodyEnvio))
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error creando request: " + err.Error()})
			return
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			ctx.JSON(http.StatusBadGateway, gin.H{"error": "error llamando al banco: " + err.Error()})
			return
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error leyendo respuesta del banco"})
			return
		}

		ctx.Data(resp.StatusCode, "application/json", respBody)
	}
}
