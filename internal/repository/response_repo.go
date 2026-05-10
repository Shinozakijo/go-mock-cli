package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shinozakijo/go-mock-cli/internal/model"
)

type ResponseRepository struct {
	db *pgxpool.Pool
}

func NewResponseRepository(db *pgxpool.Pool) *ResponseRepository {
	return &ResponseRepository{db: db}
}

func (r *ResponseRepository) GetByID(ctx context.Context, id string) (*model.Response, error) {
	var res model.Response
	err := r.db.QueryRow(ctx, `
		SELECT id, route_id, name, status_code, body, headers, delay_ms, is_active, created_at, updated_at
		FROM responses
		WHERE id = $1
	`, id).Scan(
		&res.ID, &res.RouteID, &res.Name, &res.StatusCode,
		&res.Body, &res.Headers, &res.DelayMs, &res.IsActive,
		&res.CreatedAt, &res.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("response not found")
		}
		return nil, fmt.Errorf("GetByID: %w", err)
	}
	return &res, nil
}

func (r *ResponseRepository) GetByRouteID(ctx context.Context, routeID string) ([]model.Response, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, route_id, name, status_code, body, headers, delay_ms, is_active, created_at, updated_at
		FROM responses
		WHERE route_id = $1
		ORDER BY created_at DESC
	`, routeID)
	if err != nil {
		return nil, fmt.Errorf("GetByRouteID: %w", err)
	}
	defer rows.Close()

	var responses []model.Response
	for rows.Next() {
		var res model.Response
		err := rows.Scan(
			&res.ID, &res.RouteID, &res.Name, &res.StatusCode,
			&res.Body, &res.Headers, &res.DelayMs, &res.IsActive,
			&res.CreatedAt, &res.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan response: %w", err)
		}
		responses = append(responses, res)
	}
	return responses, nil
}

func (r *ResponseRepository) GetActiveByRouteID(ctx context.Context, routeID string) (*model.Response, error) {
	var res model.Response
	err := r.db.QueryRow(ctx, `
		SELECT id, route_id, name, status_code, body, headers, delay_ms, is_active, created_at, updated_at
		FROM responses
		WHERE route_id = $1 AND is_active = TRUE
		LIMIT 1
	`, routeID).Scan(
		&res.ID, &res.RouteID, &res.Name, &res.StatusCode,
		&res.Body, &res.Headers, &res.DelayMs, &res.IsActive,
		&res.CreatedAt, &res.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("active response not found")
		}
		return nil, fmt.Errorf("GetActiveByRouteID: %w", err)
	}
	return &res, nil
}

func (r *ResponseRepository) Create(ctx context.Context, routeID, name string, statusCode int, body, headers json.RawMessage, delayMs int) (*model.Response, error) {
	var res model.Response
	err := r.db.QueryRow(ctx, `
		INSERT INTO responses (route_id, name, status_code, body, headers, delay_ms, is_active, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, FALSE, $7)
		RETURNING id, route_id, name, status_code, body, headers, delay_ms, is_active, created_at, updated_at
	`, routeID, name, statusCode, body, headers, delayMs, time.Now()).Scan(
		&res.ID, &res.RouteID, &res.Name, &res.StatusCode,
		&res.Body, &res.Headers, &res.DelayMs, &res.IsActive,
		&res.CreatedAt, &res.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("Create response: %w", err)
	}
	return &res, nil
}

func (r *ResponseRepository) SetActive(ctx context.Context, routeID, responseID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		UPDATE responses SET is_active = FALSE, updated_at = $1
		WHERE route_id = $2
	`, time.Now(), routeID)
	if err != nil {
		return fmt.Errorf("deactivate responses: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE responses SET is_active = TRUE, updated_at = $1
		WHERE id = $2
	`, time.Now(), responseID)
	if err != nil {
		return fmt.Errorf("activate response: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *ResponseRepository) UpdateBody(ctx context.Context, id string, body json.RawMessage) error {
	_, err := r.db.Exec(ctx, `
		UPDATE responses
		SET body = $1, updated_at = $2
		WHERE id = $3
	`, body, time.Now(), id)
	if err != nil {
		return fmt.Errorf("UpdateBody: %w", err)
	}
	return nil
}

func (r *ResponseRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM responses WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("Delete response: %w", err)
	}
	return nil
}