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

type OrganizationRepository struct {
	pool *pgxpool.Pool
}

// NewOrganizationRepository menginisialisasi repository Organization berbasis PostgreSQL.
// Parameter pool merupakan pointer *pgxpool.Pool aktif untuk eksekusi query database.
// Mengembalikan pointer *OrganizationRepository yang mengimplementasikan domain.OrganizationRepository.
func NewOrganizationRepository(pool *pgxpool.Pool) *OrganizationRepository {
	return &OrganizationRepository{pool: pool}
}

// Create menyimpan data entitas Organization baru ke dalam tabel organizations.
// Parameter ctx merupakan konteks eksekusi query database.
// Parameter org merupakan pointer *domain.Organization yang akan disimpan.
// Mengembalikan error jika terjadi kegagalan eksekusi query atau duplikasi slug organisasi.
func (r *OrganizationRepository) Create(ctx context.Context, org *domain.Organization) error {
	query := `
		INSERT INTO organizations (id, name, slug, tier, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at;
	`

	if org.ID == uuid.Nil {
		org.ID = uuid.New()
	}

	err := r.pool.QueryRow(
		ctx,
		query,
		org.ID,
		org.Name,
		org.Slug,
		org.Tier,
		org.CreatedAt,
		org.UpdatedAt,
	).Scan(&org.ID, &org.CreatedAt, &org.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrConflict
		}
		return fmt.Errorf("gagal menyimpan organisasi: %w", err)
	}

	return nil
}

// GetByID mengambil data Organization dari tabel organizations berdasarkan identifier UUID.
// Parameter ctx merupakan konteks eksekusi query database.
// Parameter id merupakan UUID organisasi yang dicari.
// Mengembalikan pointer *domain.Organization jika ditemukan dan domain.ErrNotFound jika data tidak ada.
func (r *OrganizationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	query := `
		SELECT id, name, slug, tier, created_at, updated_at
		FROM organizations
		WHERE id = $1;
	`

	var org domain.Organization
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&org.ID,
		&org.Name,
		&org.Slug,
		&org.Tier,
		&org.CreatedAt,
		&org.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("gagal mengambil organisasi berdasarkan id: %w", err)
	}

	return &org, nil
}

// GetBySlug mengambil data Organization dari tabel organizations berdasarkan slug unik.
// Parameter ctx merupakan konteks eksekusi query database.
// Parameter slug merupakan identifier slug URL organisasi.
// Mengembalikan pointer *domain.Organization jika ditemukan dan domain.ErrNotFound jika data tidak ada.
func (r *OrganizationRepository) GetBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	query := `
		SELECT id, name, slug, tier, created_at, updated_at
		FROM organizations
		WHERE LOWER(slug) = LOWER($1);
	`

	var org domain.Organization
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&org.ID,
		&org.Name,
		&org.Slug,
		&org.Tier,
		&org.CreatedAt,
		&org.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("gagal mengambil organisasi berdasarkan slug: %w", err)
	}

	return &org, nil
}

// ListByUser mengambil seluruh daftar Organization yang diikuti oleh seorang pengguna.
// Parameter ctx merupakan konteks eksekusi query database.
// Parameter userID merupakan UUID pengguna yang menjadi anggota organisasi.
// Mengembalikan slice []domain.Organization dan error jika terjadi kegagalan query.
func (r *OrganizationRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Organization, error) {
	query := `
		SELECT o.id, o.name, o.slug, o.tier, o.created_at, o.updated_at
		FROM organizations o
		INNER JOIN organization_members om ON o.id = om.organization_id
		WHERE om.user_id = $1
		ORDER BY o.created_at DESC;
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar organisasi pengguna: %w", err)
	}
	defer rows.Close()

	var orgs []domain.Organization
	for rows.Next() {
		var org domain.Organization
		if err := rows.Scan(&org.ID, &org.Name, &org.Slug, &org.Tier, &org.CreatedAt, &org.UpdatedAt); err != nil {
			return nil, fmt.Errorf("gagal memindai baris organisasi: %w", err)
		}
		orgs = append(orgs, org)
	}

	return orgs, nil
}

// Update memperbarui data atribut organisasi pada tabel organizations.
// Parameter ctx merupakan konteks eksekusi query database.
// Parameter org merupakan pointer *domain.Organization dengan data terbaru.
// Mengembalikan error jika terjadi kegagalan query atau organisasi tidak ditemukan.
func (r *OrganizationRepository) Update(ctx context.Context, org *domain.Organization) error {
	query := `
		UPDATE organizations
		SET name = $2, slug = $3, tier = $4
		WHERE id = $1;
	`

	cmdTag, err := r.pool.Exec(ctx, query, org.ID, org.Name, org.Slug, org.Tier)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrConflict
		}
		return fmt.Errorf("gagal memperbarui organisasi: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// Delete menghapus organisasi dari tabel organizations berdasarkan identifier UUID.
// Parameter ctx merupakan konteks eksekusi query database.
// Parameter id merupakan UUID organisasi yang akan dihapus.
// Mengembalikan error jika terjadi kegagalan query atau organisasi tidak ditemukan.
func (r *OrganizationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM organizations WHERE id = $1;`

	cmdTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("gagal menghapus organisasi: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// AddMember menambahkan relasi keanggotaan pengguna ke dalam organisasi.
// Parameter ctx merupakan konteks eksekusi query database.
// Parameter member merupakan pointer *domain.OrganizationMember yang akan disimpan.
// Mengembalikan error jika terjadi kegagalan query atau anggota sudah terdaftar di organisasi tersebut.
func (r *OrganizationRepository) AddMember(ctx context.Context, member *domain.OrganizationMember) error {
	query := `
		INSERT INTO organization_members (id, organization_id, user_id, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at;
	`

	if member.ID == uuid.Nil {
		member.ID = uuid.New()
	}

	err := r.pool.QueryRow(
		ctx,
		query,
		member.ID,
		member.OrganizationID,
		member.UserID,
		member.Role,
		member.CreatedAt,
		member.UpdatedAt,
	).Scan(&member.ID, &member.CreatedAt, &member.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrConflict
		}
		return fmt.Errorf("gagal menambahkan anggota organisasi: %w", err)
	}

	return nil
}

// GetMember mengambil detail keanggotaan pengguna tertentu dalam suatu organisasi.
// Parameter ctx merupakan konteks eksekusi query database.
// Parameter orgID merupakan UUID organisasi.
// Parameter userID merupakan UUID pengguna yang diperiksa.
// Mengembalikan pointer *domain.OrganizationMember jika ditemukan dan domain.ErrNotFound jika pengguna bukan anggota organisasi.
func (r *OrganizationRepository) GetMember(ctx context.Context, orgID, userID uuid.UUID) (*domain.OrganizationMember, error) {
	query := `
		SELECT id, organization_id, user_id, role, created_at, updated_at
		FROM organization_members
		WHERE organization_id = $1 AND user_id = $2;
	`

	var member domain.OrganizationMember
	err := r.pool.QueryRow(ctx, query, orgID, userID).Scan(
		&member.ID,
		&member.OrganizationID,
		&member.UserID,
		&member.Role,
		&member.CreatedAt,
		&member.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("gagal mengambil data anggota organisasi: %w", err)
	}

	return &member, nil
}

// ListMembers mengambil seluruh daftar anggota dalam suatu organisasi beserta data profil penggunanya.
// Parameter ctx merupakan konteks eksekusi query database.
// Parameter orgID merupakan UUID organisasi yang diperiksa.
// Mengembalikan slice []domain.OrganizationMember dan error jika query gagal.
func (r *OrganizationRepository) ListMembers(ctx context.Context, orgID uuid.UUID) ([]domain.OrganizationMember, error) {
	query := `
		SELECT om.id, om.organization_id, om.user_id, om.role, om.created_at, om.updated_at,
		       u.id, u.email, u.full_name, u.avatar_url, u.is_active, u.created_at, u.updated_at
		FROM organization_members om
		INNER JOIN users u ON om.user_id = u.id
		WHERE om.organization_id = $1
		ORDER BY om.created_at ASC;
	`

	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar anggota organisasi: %w", err)
	}
	defer rows.Close()

	var members []domain.OrganizationMember
	for rows.Next() {
		var om domain.OrganizationMember
		var u domain.User
		if err := rows.Scan(
			&om.ID, &om.OrganizationID, &om.UserID, &om.Role, &om.CreatedAt, &om.UpdatedAt,
			&u.ID, &u.Email, &u.FullName, &u.AvatarURL, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("gagal memindai data anggota organisasi: %w", err)
		}
		om.User = &u
		members = append(members, om)
	}

	return members, nil
}

// UpdateMemberRole memperbarui hak akses peran anggota dalam suatu organisasi.
// Parameter ctx merupakan konteks eksekusi query database.
// Parameter orgID merupakan UUID organisasi.
// Parameter userID merupakan UUID pengguna yang perannya diubah.
// Parameter role merupakan peran baru yang diberikan (owner, admin, member, viewer).
// Mengembalikan error jika terjadi kegagalan query atau anggota tidak ditemukan.
func (r *OrganizationRepository) UpdateMemberRole(ctx context.Context, orgID, userID uuid.UUID, role domain.OrganizationRole) error {
	query := `
		UPDATE organization_members
		SET role = $3
		WHERE organization_id = $1 AND user_id = $2;
	`

	cmdTag, err := r.pool.Exec(ctx, query, orgID, userID, role)
	if err != nil {
		return fmt.Errorf("gagal memperbarui peran anggota organisasi: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// RemoveMember menghapus keanggotaan pengguna dari suatu organisasi.
// Parameter ctx merupakan konteks eksekusi query database.
// Parameter orgID merupakan UUID organisasi.
// Parameter userID merupakan UUID pengguna yang akan dikeluarkan.
// Mengembalikan error jika terjadi kegagalan query atau anggota tidak ditemukan.
func (r *OrganizationRepository) RemoveMember(ctx context.Context, orgID, userID uuid.UUID) error {
	query := `
		DELETE FROM organization_members
		WHERE organization_id = $1 AND user_id = $2;
	`

	cmdTag, err := r.pool.Exec(ctx, query, orgID, userID)
	if err != nil {
		return fmt.Errorf("gagal mengeluarkan anggota dari organisasi: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// CreateInvitation menyimpan data undangan baru ke dalam tabel organization_invitations.
func (r *OrganizationRepository) CreateInvitation(ctx context.Context, inv *domain.OrganizationInvitation) error {
	query := `
		INSERT INTO organization_invitations (id, organization_id, email, role, token, invited_by, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at;
	`

	if inv.ID == uuid.Nil {
		inv.ID = uuid.New()
	}

	err := r.pool.QueryRow(
		ctx,
		query,
		inv.ID,
		inv.OrganizationID,
		inv.Email,
		inv.Role,
		inv.Token,
		inv.InvitedBy,
		inv.ExpiresAt,
		inv.CreatedAt,
	).Scan(&inv.ID, &inv.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrConflict
		}
		return fmt.Errorf("gagal membuat undangan organisasi: %w", err)
	}

	return nil
}

// GetInvitationByToken mengambil data undangan berdasarkan token rahasia.
func (r *OrganizationRepository) GetInvitationByToken(ctx context.Context, token string) (*domain.OrganizationInvitation, error) {
	query := `
		SELECT id, organization_id, email, role, token, invited_by, expires_at, created_at
		FROM organization_invitations
		WHERE token = $1;
	`

	var inv domain.OrganizationInvitation
	err := r.pool.QueryRow(ctx, query, token).Scan(
		&inv.ID,
		&inv.OrganizationID,
		&inv.Email,
		&inv.Role,
		&inv.Token,
		&inv.InvitedBy,
		&inv.ExpiresAt,
		&inv.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("gagal mengambil undangan berdasarkan token: %w", err)
	}

	return &inv, nil
}

// ListInvitations mengambil seluruh daftar undangan aktif dalam organisasi.
func (r *OrganizationRepository) ListInvitations(ctx context.Context, orgID uuid.UUID) ([]domain.OrganizationInvitation, error) {
	query := `
		SELECT id, organization_id, email, role, token, invited_by, expires_at, created_at
		FROM organization_invitations
		WHERE organization_id = $1 AND expires_at > CURRENT_TIMESTAMP
		ORDER BY created_at DESC;
	`

	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar undangan organisasi: %w", err)
	}
	defer rows.Close()

	var invitations []domain.OrganizationInvitation
	for rows.Next() {
		var inv domain.OrganizationInvitation
		if err := rows.Scan(
			&inv.ID,
			&inv.OrganizationID,
			&inv.Email,
			&inv.Role,
			&inv.Token,
			&inv.InvitedBy,
			&inv.ExpiresAt,
			&inv.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("gagal memindai data undangan: %w", err)
		}
		invitations = append(invitations, inv)
	}

	return invitations, nil
}

// DeleteInvitation menghapus undangan organisasi berdasarkan ID.
func (r *OrganizationRepository) DeleteInvitation(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM organization_invitations
		WHERE id = $1;
	`

	cmdTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("gagal menghapus undangan organisasi: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

