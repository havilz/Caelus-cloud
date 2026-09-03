package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrganizationRepository struct {
	pool *pgxpool.Pool
}

func NewOrganizationRepository(pool *pgxpool.Pool) *OrganizationRepository {
	return &OrganizationRepository{pool: pool}
}

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
