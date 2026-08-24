package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

type VolumeRepository struct {
	pool *pgxpool.Pool
}

func NewVolumeRepository(pool *pgxpool.Pool) *VolumeRepository {
	return &VolumeRepository{pool: pool}
}

func (r *VolumeRepository) CreateVolume(ctx context.Context, vol *domain.Volume) error {
	query := `
		INSERT INTO volumes (
			id, organization_id, server_id, name, size_gb, type, fs_type, mount_path,
			status, attached_container_name, iops, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		) RETURNING created_at, updated_at;
	`
	if vol.ID == uuid.Nil {
		vol.ID = uuid.New()
	}
	now := time.Now().UTC()
	vol.CreatedAt = now
	vol.UpdatedAt = now

	return r.pool.QueryRow(
		ctx,
		query,
		vol.ID,
		vol.OrganizationID,
		vol.ServerID,
		vol.Name,
		vol.SizeGB,
		vol.Type,
		vol.FSType,
		vol.MountPath,
		vol.Status,
		vol.AttachedContainerName,
		vol.IOPS,
		vol.CreatedAt,
		vol.UpdatedAt,
	).Scan(&vol.CreatedAt, &vol.UpdatedAt)
}

func (r *VolumeRepository) GetVolumeByID(ctx context.Context, id uuid.UUID) (*domain.Volume, error) {
	query := `
		SELECT id, organization_id, server_id, name, size_gb, type, fs_type, mount_path,
		       status, COALESCE(attached_container_name, ''), iops, created_at, updated_at
		FROM volumes
		WHERE id = $1;
	`
	var v domain.Volume
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&v.ID,
		&v.OrganizationID,
		&v.ServerID,
		&v.Name,
		&v.SizeGB,
		&v.Type,
		&v.FSType,
		&v.MountPath,
		&v.Status,
		&v.AttachedContainerName,
		&v.IOPS,
		&v.CreatedAt,
		&v.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("volume not found")
		}
		return nil, err
	}
	return &v, nil
}

func (r *VolumeRepository) ListVolumesByOrg(ctx context.Context, orgID uuid.UUID) ([]domain.Volume, error) {
	query := `
		SELECT id, organization_id, server_id, name, size_gb, type, fs_type, mount_path,
		       status, COALESCE(attached_container_name, ''), iops, created_at, updated_at
		FROM volumes
		WHERE organization_id = $1
		ORDER BY created_at DESC;
	`
	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]domain.Volume, 0)
	for rows.Next() {
		var v domain.Volume
		if err := rows.Scan(
			&v.ID,
			&v.OrganizationID,
			&v.ServerID,
			&v.Name,
			&v.SizeGB,
			&v.Type,
			&v.FSType,
			&v.MountPath,
			&v.Status,
			&v.AttachedContainerName,
			&v.IOPS,
			&v.CreatedAt,
			&v.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, v)
	}
	return list, nil
}

func (r *VolumeRepository) ListVolumesByServer(ctx context.Context, serverID uuid.UUID) ([]domain.Volume, error) {
	query := `
		SELECT id, organization_id, server_id, name, size_gb, type, fs_type, mount_path,
		       status, COALESCE(attached_container_name, ''), iops, created_at, updated_at
		FROM volumes
		WHERE server_id = $1
		ORDER BY created_at DESC;
	`
	rows, err := r.pool.Query(ctx, query, serverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]domain.Volume, 0)
	for rows.Next() {
		var v domain.Volume
		if err := rows.Scan(
			&v.ID,
			&v.OrganizationID,
			&v.ServerID,
			&v.Name,
			&v.SizeGB,
			&v.Type,
			&v.FSType,
			&v.MountPath,
			&v.Status,
			&v.AttachedContainerName,
			&v.IOPS,
			&v.CreatedAt,
			&v.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, v)
	}
	return list, nil
}

func (r *VolumeRepository) UpdateVolumeStatus(ctx context.Context, id uuid.UUID, status domain.VolumeStatus, attachedContainer string) error {
	query := `
		UPDATE volumes
		SET status = $2, attached_container_name = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1;
	`
	_, err := r.pool.Exec(ctx, query, id, status, attachedContainer)
	return err
}

func (r *VolumeRepository) UpdateVolumeSize(ctx context.Context, id uuid.UUID, newSizeGB int) error {
	query := `
		UPDATE volumes
		SET size_gb = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1;
	`
	_, err := r.pool.Exec(ctx, query, id, newSizeGB)
	return err
}

func (r *VolumeRepository) DeleteVolume(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM volumes WHERE id = $1;`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}
