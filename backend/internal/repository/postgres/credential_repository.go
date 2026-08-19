package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

type CredentialRepository struct {
	pool *pgxpool.Pool
}

// NewCredentialRepository menginisialisasi repository Kredensial Provider berbasis database PostgreSQL.
// Parameter pool merupakan instance koneksi pool *pgxpool.Pool.
// Mengembalikan pointer *CredentialRepository yang mengimplementasikan domain.CredentialRepository.
func NewCredentialRepository(pool *pgxpool.Pool) *CredentialRepository {
	return &CredentialRepository{pool: pool}
}

// Create menyimpan kredensial provider baru ke dalam tabel credentials.
// Parameter ctx merupakan konteks eksekusi query database.
// Parameter cred merupakan pointer entitas *domain.Credential yang akan disimpan.
// Mengembalikan error jika terjadi kegagalan query atau pelanggaran integritas constraint.
func (r *CredentialRepository) Create(ctx context.Context, cred *domain.Credential) error {
	query := `
		INSERT INTO credentials (id, organization_id, provider_id, name, encrypted_api_key, encrypted_api_secret, encrypted_ssh_key, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at;
	`

	if cred.ID == uuid.Nil {
		cred.ID = uuid.New()
	}

	var metadataJSON []byte
	var err error
	if cred.Metadata != nil {
		metadataJSON, err = json.Marshal(cred.Metadata)
		if err != nil {
			return fmt.Errorf("gagal serialisasi metadata kredensial: %w", err)
		}
	}

	err = r.pool.QueryRow(
		ctx,
		query,
		cred.ID,
		cred.OrganizationID,
		cred.ProviderID,
		cred.Name,
		cred.EncryptedAPIKey,
		cred.EncryptedAPISecret,
		cred.EncryptedSSHKey,
		metadataJSON,
		cred.CreatedAt,
		cred.UpdatedAt,
	).Scan(&cred.ID, &cred.CreatedAt, &cred.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrConflict
		}
		return fmt.Errorf("gagal menyimpan kredensial provider: %w", err)
	}

	return nil
}

// GetByID mengambil data kredensial provider beserta relasi Provider berdasarkan identifier UUID.
// Parameter ctx merupakan konteks eksekusi query database.
// Parameter id merupakan UUID kredensial yang dicari.
// Mengembalikan pointer *domain.Credential jika ditemukan dan domain.ErrNotFound jika data tidak ada.
func (r *CredentialRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Credential, error) {
	query := `
		SELECT c.id, c.organization_id, c.provider_id, c.name, c.encrypted_api_key, c.encrypted_api_secret, c.encrypted_ssh_key, c.metadata, c.created_at, c.updated_at,
		       p.id, p.name, p.slug, p.is_active, p.created_at
		FROM credentials c
		INNER JOIN providers p ON c.provider_id = p.id
		WHERE c.id = $1;
	`

	var c domain.Credential
	var p domain.Provider
	var metadataJSON []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&c.ID,
		&c.OrganizationID,
		&c.ProviderID,
		&c.Name,
		&c.EncryptedAPIKey,
		&c.EncryptedAPISecret,
		&c.EncryptedSSHKey,
		&metadataJSON,
		&c.CreatedAt,
		&c.UpdatedAt,
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
		return nil, fmt.Errorf("gagal mengambil kredensial berdasarkan id: %w", err)
	}

	if len(metadataJSON) > 0 {
		_ = json.Unmarshal(metadataJSON, &c.Metadata)
	}
	c.Provider = &p

	return &c, nil
}

// ListByOrg mengambil seluruh daftar kredensial provider yang dimiliki oleh suatu organisasi.
// Parameter ctx merupakan konteks eksekusi query database.
// Parameter orgID merupakan UUID organisasi pemilik kredensial.
// Mengembalikan slice []domain.Credential dan error jika query gagal.
func (r *CredentialRepository) ListByOrg(ctx context.Context, orgID uuid.UUID) ([]domain.Credential, error) {
	query := `
		SELECT c.id, c.organization_id, c.provider_id, c.name, c.encrypted_api_key, c.encrypted_api_secret, c.encrypted_ssh_key, c.metadata, c.created_at, c.updated_at,
		       p.id, p.name, p.slug, p.is_active, p.created_at
		FROM credentials c
		INNER JOIN providers p ON c.provider_id = p.id
		WHERE c.organization_id = $1
		ORDER BY c.created_at DESC;
	`

	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar kredensial organisasi: %w", err)
	}
	defer rows.Close()

	var creds []domain.Credential
	for rows.Next() {
		var c domain.Credential
		var p domain.Provider
		var metadataJSON []byte

		if err := rows.Scan(
			&c.ID,
			&c.OrganizationID,
			&c.ProviderID,
			&c.Name,
			&c.EncryptedAPIKey,
			&c.EncryptedAPISecret,
			&c.EncryptedSSHKey,
			&metadataJSON,
			&c.CreatedAt,
			&c.UpdatedAt,
			&p.ID,
			&p.Name,
			&p.Slug,
			&p.IsActive,
			&p.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("gagal memindai baris kredensial: %w", err)
		}

		if len(metadataJSON) > 0 {
			_ = json.Unmarshal(metadataJSON, &c.Metadata)
		}
		c.Provider = &p
		creds = append(creds, c)
	}

	return creds, nil
}

// Update memperbarui data atribut kredensial pada tabel credentials.
// Parameter ctx merupakan konteks eksekusi query database.
// Parameter cred merupakan pointer *domain.Credential dengan data terbaru.
// Mengembalikan error jika terjadi kegagalan query atau kredensial tidak ditemukan.
func (r *CredentialRepository) Update(ctx context.Context, cred *domain.Credential) error {
	query := `
		UPDATE credentials
		SET name = $2, encrypted_api_key = $3, encrypted_api_secret = $4, encrypted_ssh_key = $5, metadata = $6
		WHERE id = $1;
	`

	var metadataJSON []byte
	var err error
	if cred.Metadata != nil {
		metadataJSON, err = json.Marshal(cred.Metadata)
		if err != nil {
			return fmt.Errorf("gagal serialisasi metadata kredensial: %w", err)
		}
	}

	cmdTag, err := r.pool.Exec(ctx, query, cred.ID, cred.Name, cred.EncryptedAPIKey, cred.EncryptedAPISecret, cred.EncryptedSSHKey, metadataJSON)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrConflict
		}
		return fmt.Errorf("gagal memperbarui kredensial: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// Delete menghapus data kredensial dari tabel credentials berdasarkan identifier UUID.
// Parameter ctx merupakan konteks eksekusi query database.
// Parameter id merupakan UUID kredensial yang akan dihapus.
// Mengembalikan error jika terjadi kegagalan query atau kredensial tidak ditemukan.
func (r *CredentialRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM credentials WHERE id = $1;`

	cmdTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("gagal menghapus kredensial: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
