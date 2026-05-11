package tui

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/shinozakijo/go-mock-cli/internal/repository"
	"github.com/shinozakijo/go-mock-cli/internal/server"
	"github.com/shinozakijo/go-mock-cli/internal/service"
)

type ServerManager struct {
	mu           sync.Mutex
	running      bool
	port         string
	httpServer   *http.Server
	routeRepo    *repository.RouteRepository
	responseRepo *repository.ResponseRepository
}

func NewServerManager(
	port string,
	routeRepo *repository.RouteRepository,
	responseRepo *repository.ResponseRepository,
) *ServerManager {
	return &ServerManager{
		port:         port,
		routeRepo:    routeRepo,
		responseRepo: responseRepo,
	}
}

func (sm *ServerManager) Start() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.running {
		return fmt.Errorf("server already running")
	}

	mockService := service.NewMockService(sm.routeRepo, sm.responseRepo)
	handler := server.NewHandler(mockService)
	ginEngine := server.NewEngine(handler)

	sm.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%s", sm.port),
		Handler: ginEngine,
	}

	go func() {
		sm.httpServer.ListenAndServe()
	}()

	sm.running = true
	return nil
}

func (sm *ServerManager) Stop() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.running {
		return fmt.Errorf("server not running")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := sm.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	sm.running = false
	sm.httpServer = nil
	return nil
}

func (sm *ServerManager) IsRunning() bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.running
}

func (sm *ServerManager) Port() string {
	return sm.port
}
