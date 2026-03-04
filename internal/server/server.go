package server

import (
	"net"
	"net/http"
	"time"
	"v2/internal/api"
	"v2/internal/handler"

	"github.com/gin-gonic/gin"
)

type Server struct {
	Port   string
	Engine *gin.Engine
}

func NewServer(port string, cfg handler.Config) *Server {
	engine := gin.Default()
	registerRoutes(engine, cfg)
	return &Server{
		Port:   port,
		Engine: engine,
	}
}

func (s *Server) Run() error {
	listener, err := net.Listen("tcp", s.Port)
	if err != nil {
		return err
	}
	srv := &http.Server{
		Handler:      s.Engine,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  10 * time.Second,
	}
	return srv.Serve(listener)
}

func registerRoutes(r *gin.Engine, cfg handler.Config) {
	r.GET("/api/v1/health", api.CheckHealth(cfg))
	r.POST("/simf/api/v1/config/solest", api.SolestConfig(cfg))
}
