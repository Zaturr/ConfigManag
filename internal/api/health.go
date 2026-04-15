package api

import (
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"v2/internal/handler"

	"github.com/gin-gonic/gin"
)

// Puerto y path para el health check (misma base que solest: IP del config + 8082).
// Endpoint de health hardcodeado.
const (
	healthCheckPort = "8082"
	healthCheckPath = "/estatusrest"
	healthTimeout   = 12 * time.Second
)

type HealthResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

// RunHealthCheck ejecuta el health check sobre cfg y devuelve el resultado.
// Usa la misma lógica que solestconfig para la IP: banco.IP del config, puerto 8082,
// y un endpoint de health hardcodeado (no el endpoint del banco).
func RunHealthCheck(cfg handler.Config) HealthResponse {
	resp := HealthResponse{
		Status: "UP",
		Checks: make(map[string]string),
	}
	client := &http.Client{Timeout: healthTimeout}
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, banco := range cfg {
		banco := banco
		wg.Add(1)
		go func() {
			defer wg.Done()
			ip := strings.TrimSpace(banco.IP)
			if ip == "" {
				mu.Lock()
				resp.Checks[banco.Nombre] = "sin IP en configuración"
				resp.Status = "Error"
				mu.Unlock()
				handler.LogHealthCheck(
					banco.Nombre,
					"",
					"sin IP en configuración",
					-1,
					"",
					"",
				)
				return
			}
			path := healthCheckPath
			if !strings.HasPrefix(path, "/") {
				path = "/" + path
			}
			urlHealth := "http://" + ip + ":" + healthCheckPort + path
			r, err := client.Get(urlHealth)
			if err != nil {
				rawErr := err.Error()
				statusCode := -1
				if strings.Contains(strings.ToLower(rawErr), "timeout") ||
					strings.Contains(strings.ToLower(rawErr), "deadline exceeded") {
					statusCode = http.StatusRequestTimeout
				}
				mu.Lock()
				resp.Checks[banco.Nombre] = rawErr
				resp.Status = "Error"
				mu.Unlock()
				handler.LogHealthCheck(
					banco.Nombre,
					urlHealth,
					rawErr,
					statusCode,
					"",
					"",
				)
				return
			}
			bodyBytes, readErr := io.ReadAll(r.Body)
			r.Body.Close()
			rawBody := strings.TrimSpace(string(bodyBytes))
			received := strings.TrimSpace(r.Status)
			if rawBody != "" {
				received = rawBody
			}
			if received == "" {
				received = "respuesta vacía"
			}
			if readErr != nil {
				received = received + " | error leyendo body: " + readErr.Error()
			}
			mu.Lock()
			resp.Checks[banco.Nombre] = received
			mu.Unlock()
			handler.LogHealthCheck(
				banco.Nombre,
				urlHealth,
				received,
				r.StatusCode,
				r.Status,
				rawBody,
			)
		}()
	}
	wg.Wait()
	return resp
}

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
