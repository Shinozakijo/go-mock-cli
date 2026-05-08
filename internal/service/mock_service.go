package service

import (
	"context"
	"fmt"

	"github.com/shinozakijo/go-mock-cli/internal/model"
	"github.com/shinozakijo/go-mock-cli/internal/repository"
)

type MockService struct {
	routeRepo    *repository.RouteRepository
	responseRepo *repository.ResponseRepository
}

func NewMockService(
	routeRepo *repository.RouteRepository,
	responseRepo *repository.ResponseRepository,
) *MockService {
	return &MockService{
		routeRepo:    routeRepo,
		responseRepo: responseRepo,
	}
}

func (s *MockService) FindMockResponse(
	ctx context.Context,
	method string,
	path string,
) (*model.Route, *model.Response, error) {
	route, err := s.routeRepo.GetByMethodAndPath(ctx, method, path)
	if err != nil {
		return nil, nil, fmt.Errorf("route not found: %w", err)
	}

	response, err := s.responseRepo.GetActiveByRouteID(ctx, route.ID)
	if err != nil {
		return route, nil, fmt.Errorf("active response not found: %w", err)
	}

	return route, response, nil
}