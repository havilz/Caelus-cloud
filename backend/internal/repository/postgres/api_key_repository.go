package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

type APIKeyRepository struct {
	pool *pgxpool.Pool
}

func NewAPIKeyRepository(pool *pgxpool.Pool) *APIKeyRepository {
	return &APIKeyRepository{pool: pool}
}

func (r *APIKeyRepository) Create(ctx context.Context, apiKey *domain.APIKey) error {
	query := `
		INSERT INTO api_keys (id, organization_id, user_id, name, key_prefix, key_hash, scopes, expires_at, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at;
	`

	if apiKey.ID == uuid.Nil {
		apiKey.ID = uuid.New()
	}

	err := r.pool.QueryRow(
		ctx,
		query,
		apiKey.ID,
		apiKey.OrganizationID,
		apiKey.UserID,
		apiKey.Name,
		apiKey.KeyPrefix,
		apiKey.KeyHash,
		apiKey.Scopes,
		apiKey.ExpiresAt,
		apiKey.IsActive,
		apiKey.CreatedAt,
		apiKey.UpdatedAt,
	).Scan(&apiKey.ID, &apiKey.CreatedAt, &apiKey.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrConflict
		}
		return fmt.Errorf("gagal membuat API key: %w", err)
	}

	return nil
}

func (r *APIKeyRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.APIKey, error) {
	query := `
		SELECT id, organization_id, user_id, name, key_prefix, key_hash, scopes, last_used_at, expires_at, is_active, created_at, updated_at
		FROM api_keys
		WHERE id = $1;
	`

	var k domain.APIKey
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&k.ID,
		&k.OrganizationID,
		&k.UserID,
		&k.Name,
		&k.KeyPrefix,
		&k.KeyHash,
		&k.Scopes,
		&k.LastUsedAt,
		&k.ExpiresAt,
		&k.IsActive,
		&k.CreatedAt,
		&k.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("gagal mengambil API key berdasarkan ID: %w", err)
	}

	return &k, nil
}

func (r *APIKeyRepository) GetByHash(ctx context.Context, keyHash string) (*domain.APIKey, error) {
	query := `
		SELECT id, organization_id, user_id, name, key_prefix, key_hash, scopes, last_used_at, expires_at, is_active, created_at, updated_at
		FROM api_keys
		WHERE key_hash = $1 AND is_active = true;
	`

	var k domain.APIKey
	err := r.pool.QueryRow(ctx, query, keyHash).Scan(
		&k.ID,
		&k.OrganizationID,
		&k.UserID,
		&k.Name,
		&k.KeyPrefix,
		&k.KeyHash,
		&k.Scopes,
		&k.LastUsedAt,
		&k.ExpiresAt,
		&k.IsActive,
		&k.CreatedAt,
		&k.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("gagal mengambil API key berdasarkan hash: %w", err)
	}

	return &k, nil
}

func (r *APIKeyRepository) ListByOrg(ctx context.Context, orgID uuid.UUID) ([]domain.APIKey, error) {
	query := `
		SELECT id, organization_id, user_id, name, key_prefix, key_hash, scopes, last_used_at, expires_at, is_active, created_at, updated_at
		FROM api_keys
		WHERE organization_id = $1
		ORDER BY created_at DESC;
	`

	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar API keys: %w", err)
	}
	defer rows.Close()

	var keys []domain.APIKey
	for rows.Next() {
		var k domain.APIKey
		if err := rows.Scan(
			&k.ID,
			&k.OrganizationID,
			&k.UserID,
			&k.Name,
			&k.KeyPrefix,
			&k.KeyHash,
			&k.Scopes,
			&k.LastUsedAt,
			&k.ExpiresAt,
			&k.IsActive,
			&k.CreatedAt,
			&k.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("gagal memindai data API key: %w", err)
		}
		keys = append(keys, k)
	}

	return keys, nil
}

func (r *APIKeyRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID, lastUsed time.Time) error {
	query := `
		UPDATE api_keys
		SET last_used_at = $2
		WHERE id = $1;
	`

	_, err := r.pool.Exec(ctx, query, id, lastUsed)
	return err
}

func (r *APIKeyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM api_keys
		WHERE id = $1;
	`

	cmdTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("gagal menghapus API key: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
