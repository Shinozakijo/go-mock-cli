package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shinozakijo/go-mock-cli/internal/service"
)

type Handler struct {
	mockService *service.MockService
}

func NewHandler(mockService *service.MockService) *Handler {
	return &Handler{
		mockService: mockService,
	}
}

func (h *Handler) HandleMock(c *gin.Context) {
	method := c.Request.Method
	path := c.Request.URL.Path

	_, response, err := h.mockService.FindMockResponse(c.Request.Context(), method, path)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "mock route not found",
			"method":  method,
			"path":    path,
			"details": err.Error(),
		})
		return
	}

	if response.DelayMs > 0 {
		time.Sleep(time.Duration(response.DelayMs) * time.Millisecond)
	}

	applyHeaders(c, response.Headers)

	contentType := "application/json"
	c.Data(response.StatusCode, contentType, response.Body)
}