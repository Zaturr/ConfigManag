package server

import (
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
	return s.Engine.Run(s.Port)
}

func registerRoutes(r *gin.Engine, cfg handler.Config) {
	r.GET("/api/v1/health", api.CheckHealth(cfg))
}
