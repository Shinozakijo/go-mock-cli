package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shinozakijo/go-mock-cli/internal/model"
)

type RouteRepository struct {
	db *pgxpool.Pool
}

func NewRouteRepository(db *pgxpool.Pool) *RouteRepository {
	return &RouteRepository{db: db}
}

// ดึง routes ทั้งหมด
func (r *RouteRepository) GetAll(ctx context.Context) ([]model.Route, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, method, path, description, created_at, updated_at
		FROM routes
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("GetAll routes: %w", err)
	}
	defer rows.Close()

	var routes []model.Route
	for rows.Next() {
		var route model.Route
		err := rows.Scan(
			&route.ID, &route.Method, &route.Path,
			&route.Description, &route.CreatedAt, &route.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan route: %w", err)
		}
		routes = append(routes, route)
	}
	return routes, nil
}

// ดึง route ด้วย method + path (ใช้ตอน mock server match request)
func (r *RouteRepository) GetByMethodAndPath(ctx context.Context, method, path string) (*model.Route, error) {
	var route model.Route
	err := r.db.QueryRow(ctx, `
		SELECT id, method, path, description, created_at, updated_at
		FROM routes
		WHERE method = $1 AND path = $2
	`, method, path).Scan(
		&route.ID, &route.Method, &route.Path,
		&route.Description, &route.CreatedAt, &route.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("GetByMethodAndPath: %w", err)
	}
	return &route, nil
}

// เพิ่ม route ใหม่
func (r *RouteRepository) Create(ctx context.Context, method, path, description string) (*model.Route, error) {
	var route model.Route
	err := r.db.QueryRow(ctx, `
		INSERT INTO routes (method, path, description, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, method, path, description, created_at, updated_at
	`, method, path, description, time.Now()).Scan(
		&route.ID, &route.Method, &route.Path,
		&route.Description, &route.CreatedAt, &route.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("Create route: %w", err)
	}
	return &route, nil
}

// ลบ route
func (r *RouteRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM routes WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("Delete route: %w", err)
	}
	return nil
}
