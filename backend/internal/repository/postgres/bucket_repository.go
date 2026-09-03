package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type bucketRepository struct {
	pool *pgxpool.Pool
}

func NewBucketRepository(pool *pgxpool.Pool) domain.BucketRepository {
	return &bucketRepository{pool: pool}
}

func (r *bucketRepository) Create(ctx context.Context, bucket *domain.Bucket) error {
	query := `
		INSERT INTO buckets (id, organization_id, name, provider_type, region, is_public, versioning, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at
	`

	if bucket.ID == uuid.Nil {
		bucket.ID = uuid.New()
	}

	err := r.pool.QueryRow(ctx, query,
		bucket.ID,
		bucket.OrganizationID,
		bucket.Name,
		bucket.ProviderType,
		bucket.Region,
		bucket.IsPublic,
		bucket.Versioning,
		bucket.CreatedAt,
		bucket.UpdatedAt,
	).Scan(&bucket.CreatedAt, &bucket.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert bucket record: %w", err)
	}

	return nil
}

func (r *bucketRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Bucket, error) {
	query := `
		SELECT id, organization_id, name, provider_type, region, is_public, versioning, created_at, updated_at
		FROM buckets
		WHERE id = $1
	`

	var b domain.Bucket
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&b.ID,
		&b.OrganizationID,
		&b.Name,
		&b.ProviderType,
		&b.Region,
		&b.IsPublic,
		&b.Versioning,
		&b.CreatedAt,
		&b.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to query bucket by ID: %w", err)
	}

	return &b, nil
}

func (r *bucketRepository) GetByName(ctx context.Context, name string) (*domain.Bucket, error) {
	query := `
		SELECT id, organization_id, name, provider_type, region, is_public, versioning, created_at, updated_at
		FROM buckets
		WHERE name = $1
	`

	var b domain.Bucket
	err := r.pool.QueryRow(ctx, query, name).Scan(
		&b.ID,
		&b.OrganizationID,
		&b.Name,
		&b.ProviderType,
		&b.Region,
		&b.IsPublic,
		&b.Versioning,
		&b.CreatedAt,
		&b.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to query bucket by name: %w", err)
	}

	return &b, nil
}

func (r *bucketRepository) ListByOrgID(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]domain.Bucket, int, error) {
	countQuery := `SELECT COUNT(*) FROM buckets WHERE organization_id = $1`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, orgID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count buckets: %w", err)
	}

	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, organization_id, name, provider_type, region, is_public, versioning, created_at, updated_at
		FROM buckets
		WHERE organization_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, orgID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list buckets: %w", err)
	}
	defer rows.Close()

	buckets := make([]domain.Bucket, 0)
	for rows.Next() {
		var b domain.Bucket
		if err := rows.Scan(
			&b.ID,
			&b.OrganizationID,
			&b.Name,
			&b.ProviderType,
			&b.Region,
			&b.IsPublic,
			&b.Versioning,
			&b.CreatedAt,
			&b.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan bucket row: %w", err)
		}
		buckets = append(buckets, b)
	}

	return buckets, total, nil
}

func (r *bucketRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM buckets WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete bucket: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *bucketRepository) CountByOrgID(ctx context.Context, orgID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM buckets WHERE organization_id = $1`
	var count int
	if err := r.pool.QueryRow(ctx, query, orgID).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count org buckets: %w", err)
	}
	return count, nil
}
