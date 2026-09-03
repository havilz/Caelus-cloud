package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WebhookRepository struct {
	pool *pgxpool.Pool
}

func NewWebhookRepository(pool *pgxpool.Pool) *WebhookRepository {
	return &WebhookRepository{pool: pool}
}

func (r *WebhookRepository) Create(ctx context.Context, webhook *domain.Webhook) error {
	query := `
		INSERT INTO webhooks (id, organization_id, name, url, secret, events, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at;
	`

	if webhook.ID == uuid.Nil {
		webhook.ID = uuid.New()
	}

	err := r.pool.QueryRow(
		ctx,
		query,
		webhook.ID,
		webhook.OrganizationID,
		webhook.Name,
		webhook.URL,
		webhook.Secret,
		webhook.Events,
		webhook.IsActive,
		webhook.CreatedAt,
		webhook.UpdatedAt,
	).Scan(&webhook.ID, &webhook.CreatedAt, &webhook.UpdatedAt)

	if err != nil {
		return fmt.Errorf("gagal membuat webhook: %w", err)
	}

	return nil
}

func (r *WebhookRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Webhook, error) {
	query := `
		SELECT id, organization_id, name, url, secret, events, is_active, last_triggered_at, last_status, created_at, updated_at
		FROM webhooks
		WHERE id = $1;
	`

	var w domain.Webhook
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&w.ID,
		&w.OrganizationID,
		&w.Name,
		&w.URL,
		&w.Secret,
		&w.Events,
		&w.IsActive,
		&w.LastTriggeredAt,
		&w.LastStatus,
		&w.CreatedAt,
		&w.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("gagal mengambil webhook berdasarkan ID: %w", err)
	}

	return &w, nil
}

func (r *WebhookRepository) ListByOrg(ctx context.Context, orgID uuid.UUID) ([]domain.Webhook, error) {
	query := `
		SELECT id, organization_id, name, url, secret, events, is_active, last_triggered_at, last_status, created_at, updated_at
		FROM webhooks
		WHERE organization_id = $1
		ORDER BY created_at DESC;
	`

	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar webhook: %w", err)
	}
	defer rows.Close()

	var webhooks []domain.Webhook
	for rows.Next() {
		var w domain.Webhook
		if err := rows.Scan(
			&w.ID,
			&w.OrganizationID,
			&w.Name,
			&w.URL,
			&w.Secret,
			&w.Events,
			&w.IsActive,
			&w.LastTriggeredAt,
			&w.LastStatus,
			&w.CreatedAt,
			&w.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("gagal memindai data webhook: %w", err)
		}
		webhooks = append(webhooks, w)
	}

	return webhooks, nil
}

func (r *WebhookRepository) ListByEvent(ctx context.Context, orgID uuid.UUID, event string) ([]domain.Webhook, error) {
	query := `
		SELECT id, organization_id, name, url, secret, events, is_active, last_triggered_at, last_status, created_at, updated_at
		FROM webhooks
		WHERE organization_id = $1 AND is_active = true AND $2 = ANY(events);
	`

	rows, err := r.pool.Query(ctx, query, orgID, event)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil webhook berdasarkan event: %w", err)
	}
	defer rows.Close()

	var webhooks []domain.Webhook
	for rows.Next() {
		var w domain.Webhook
		if err := rows.Scan(
			&w.ID,
			&w.OrganizationID,
			&w.Name,
			&w.URL,
			&w.Secret,
			&w.Events,
			&w.IsActive,
			&w.LastTriggeredAt,
			&w.LastStatus,
			&w.CreatedAt,
			&w.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("gagal memindai webhook event: %w", err)
		}
		webhooks = append(webhooks, w)
	}

	return webhooks, nil
}

func (r *WebhookRepository) Update(ctx context.Context, webhook *domain.Webhook) error {
	query := `
		UPDATE webhooks
		SET name = $2, url = $3, secret = $4, events = $5, is_active = $6, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1;
	`

	cmdTag, err := r.pool.Exec(ctx, query, webhook.ID, webhook.Name, webhook.URL, webhook.Secret, webhook.Events, webhook.IsActive)
	if err != nil {
		return fmt.Errorf("gagal memperbarui webhook: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *WebhookRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status int, triggeredAt time.Time) error {
	query := `
		UPDATE webhooks
		SET last_status = $2, last_triggered_at = $3
		WHERE id = $1;
	`

	_, err := r.pool.Exec(ctx, query, id, status, triggeredAt)
	return err
}

func (r *WebhookRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM webhooks
		WHERE id = $1;
	`

	cmdTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("gagal menghapus webhook: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
