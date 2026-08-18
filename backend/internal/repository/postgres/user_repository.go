package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository menginisialisasi repository User berbasis PostgreSQL.
// Parameter pool merupakan pointer *pgxpool.Pool aktif untuk eksekusi query database.
// Mengembalikan pointer *UserRepository yang mengimplementasikan domain.UserRepository.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// Create menyimpan data entitas User baru ke dalam tabel users.
// Parameter ctx merupakan konteks eksekusi query database.
// Parameter user merupakan pointer *domain.User yang akan disimpan.
// Mengembalikan error jika terjadi kegagalan eksekusi query atau duplikasi email (domain.ErrEmailAlreadyInUse).
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, full_name, avatar_url, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at;
	`

	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	err := r.pool.QueryRow(
		ctx,
		query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.AvatarURL,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrEmailAlreadyInUse
		}
		return fmt.Errorf("gagal menyimpan pengguna: %w", err)
	}

	return nil
}

// GetByID mengambil data User dari tabel users berdasarkan identifier UUID.
// Parameter ctx merupakan konteks eksekusi query database.
// Parameter id merupakan UUID pengguna yang dicari.
// Mengembalikan pointer *domain.User jika ditemukan dan domain.ErrNotFound jika data tidak ada.
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, full_name, avatar_url, is_active, created_at, updated_at
		FROM users
		WHERE id = $1;
	`

	var user domain.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.AvatarURL,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("gagal mengambil pengguna berdasarkan id: %w", err)
	}

	return &user, nil
}

// GetByEmail mengambil data User dari tabel users berdasarkan alamat email unik.
// Parameter ctx merupakan konteks eksekusi query database.
// Parameter email merupakan alamat email pengguna yang dicari.
// Mengembalikan pointer *domain.User jika ditemukan dan domain.ErrNotFound jika data tidak ada.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, full_name, avatar_url, is_active, created_at, updated_at
		FROM users
		WHERE LOWER(email) = LOWER($1);
	`

	var user domain.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.AvatarURL,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("gagal mengambil pengguna berdasarkan email: %w", err)
	}

	return &user, nil
}

// Update memperbarui data profil pengguna pada tabel users.
// Parameter ctx merupakan konteks eksekusi query database.
// Parameter user merupakan pointer *domain.User yang memuat data terbaru.
// Mengembalikan error jika terjadi kegagalan eksekusi query atau data pengguna tidak ditemukan.
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET email = $2, password_hash = $3, full_name = $4, avatar_url = $5, is_active = $6
		WHERE id = $1;
	`

	cmdTag, err := r.pool.Exec(
		ctx,
		query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.AvatarURL,
		user.IsActive,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrEmailAlreadyInUse
		}
		return fmt.Errorf("gagal memperbarui pengguna: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// Delete menghapus data pengguna dari tabel users berdasarkan identifier UUID.
// Parameter ctx merupakan konteks eksekusi query database.
// Parameter id merupakan UUID pengguna yang akan dihapus.
// Mengembalikan error jika terjadi kegagalan eksekusi query atau data tidak ditemukan.
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1;`

	cmdTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("gagal menghapus pengguna: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
