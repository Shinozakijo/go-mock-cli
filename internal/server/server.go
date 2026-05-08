package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type Server struct {
	engine *gin.Engine
	port   string
}

func New(port string, handler *Handler) *Server {
	engine := gin.Default()

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// catch-all ทุก method ทุก path
	engine.NoRoute(handler.HandleMock)

	return &Server{
		engine: engine,
		port:   port,
	}
}

func (s *Server) Start() error {
	return s.engine.Run(fmt.Sprintf(":%s", s.port))
}