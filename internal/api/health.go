package api

import (
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
	healthCheckPath = "/simf/api/v1/health"
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
	client := &http.Client{Timeout: 5 * time.Second}
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
				return
			}
			path := healthCheckPath
			if !strings.HasPrefix(path, "/") {
				path = "/" + path
			}
			urlHealth := "http://" + ip + ":" + healthCheckPort + path
			r, err := client.Get(urlHealth)
			if err != nil {
				mu.Lock()
				resp.Checks[banco.Nombre] = "OFFLINE: " + err.Error()
				resp.Status = "Error"
				mu.Unlock()
				return
			}
			r.Body.Close()
			status := "ONLINE"
			if r.StatusCode >= 400 {
				status = "HTTP " + r.Status
			}
			mu.Lock()
			resp.Checks[banco.Nombre] = status
			if r.StatusCode >= 400 {
				resp.Status = "Error"
			}
			mu.Unlock()
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
