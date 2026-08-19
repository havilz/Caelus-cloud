package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

type ProviderRepository struct {
	pool *pgxpool.Pool
}

// NewProviderRepository menginisialisasi repository Provider berbasis database PostgreSQL.
// Parameter pool merupakan instance koneksi pool *pgxpool.Pool.
// Mengembalikan pointer *ProviderRepository yang mengimplementasikan domain.ProviderRepository.
func NewProviderRepository(pool *pgxpool.Pool) *ProviderRepository {
	return &ProviderRepository{pool: pool}
}

// GetByID mengambil data provider berdasarkan identifier UUID.
// Parameter ctx merupakan konteks eksekusi query database.
// Parameter id merupakan UUID provider yang dicari.
// Mengembalikan pointer *domain.Provider jika ditemukan dan domain.ErrNotFound jika data tidak ada.
func (r *ProviderRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Provider, error) {
	query := `
		SELECT id, name, slug, is_active, created_at
		FROM providers
		WHERE id = $1;
	`

	var p domain.Provider
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID,
		&p.Name,
		&p.Slug,
		&p.IsActive,
		&p.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("gagal mengambil provider berdasarkan id: %w", err)
	}

	return &p, nil
}

// GetBySlug mengambil data provider berdasarkan identifier slug unik (misal: "mock", "aws", "hetzner").
// Parameter ctx merupakan konteks eksekusi query database.
// Parameter slug merupakan slug unik provider.
// Mengembalikan pointer *domain.Provider jika ditemukan dan domain.ErrNotFound jika data tidak ada.
func (r *ProviderRepository) GetBySlug(ctx context.Context, slug string) (*domain.Provider, error) {
	query := `
		SELECT id, name, slug, is_active, created_at
		FROM providers
		WHERE LOWER(slug) = LOWER($1);
	`

	var p domain.Provider
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&p.ID,
		&p.Name,
		&p.Slug,
		&p.IsActive,
		&p.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("gagal mengambil provider berdasarkan slug: %w", err)
	}

	return &p, nil
}

// List mengambil seluruh daftar provider cloud yang aktif dan didukung sistem.
// Parameter ctx merupakan konteks eksekusi query database.
// Mengembalikan slice []domain.Provider dan error jika terjadi kegagalan query.
func (r *ProviderRepository) List(ctx context.Context) ([]domain.Provider, error) {
	query := `
		SELECT id, name, slug, is_active, created_at
		FROM providers
		ORDER BY created_at ASC;
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar provider: %w", err)
	}
	defer rows.Close()

	var providers []domain.Provider
	for rows.Next() {
		var p domain.Provider
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.IsActive, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("gagal memindai baris provider: %w", err)
		}
		providers = append(providers, p)
	}

	return providers, nil
}
