package api

import (
	"net"
	"net/http"
	"time"

	"v2/internal/handler"

	"github.com/gin-gonic/gin"
)

// HealthResponse es la respuesta JSON del endpoint de health.
type HealthResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

// RunHealthCheck ejecuta el health check sobre cfg y devuelve el resultado.
// Sirve tanto para el handler HTTP como para mostrarlo en CLI.
func RunHealthCheck(cfg handler.Config) HealthResponse {
	resp := HealthResponse{
		Status: "UP",
		Checks: make(map[string]string),
	}
	for _, banco := range cfg {
		addr, err := handler.ObtenerIp(banco)
		if err != nil {
			resp.Checks[banco.Nombre] = err.Error()
			resp.Status = "Error"
			continue
		}
		if err := checkTCP(addr); err != nil {
			resp.Checks[banco.Nombre] = "OFFLINE: " + err.Error()
			resp.Status = "Error"
			continue
		}
		resp.Checks[banco.Nombre] = "ONLINE"
	}
	return resp
}

// CheckHealth devuelve un handler Gin que comprueba conectividad TCP a los bancos de cfg.
// cfg es el mapa código -> banco (p. ej. config de Desarrollo o Producción).
func CheckHealth(cfg handler.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		resp := RunHealthCheck(cfg)
		status := http.StatusOK
		if resp.Status == "Error" {
			status = http.StatusServiceUnavailable
		}
		c.JSON(status, resp)
	}
}

func checkTCP(addr string) error {
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return err
	}
	_ = conn.Close()
	return nil
}
