package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Server struct {
	engine *gin.Engine
	port   string
}

// NewEngine คืน gin.Engine สำหรับ ServerManager ใน TUI ใช้
func NewEngine(handler *Handler) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	engine.NoRoute(handler.HandleMock)
	engine.NoMethod(handler.HandleMock)

	return engine
}

func New(port string, handler *Handler) *Server {
	engine := gin.Default()

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	engine.NoRoute(handler.HandleMock)
	engine.NoMethod(handler.HandleMock)

	return &Server{
		engine: engine,
		port:   port,
	}
}

func (s *Server) Start() error {
	return s.engine.Run(fmt.Sprintf(":%s", s.port))
}
