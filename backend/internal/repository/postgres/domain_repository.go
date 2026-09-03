package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DomainRepository struct {
	pool *pgxpool.Pool
}

func NewDomainRepository(pool *pgxpool.Pool) *DomainRepository {
	return &DomainRepository{pool: pool}
}

func (r *DomainRepository) Create(ctx context.Context, d *domain.CustomDomain) error {
	query := `
		INSERT INTO domains (
			id, organization_id, server_id, domain_name, target_type, target_id,
			target_port, status, verification_token, ssl_status, auto_ssl,
			cloudflare_dns_managed, cloudflare_record_id, error_message, last_checked_at,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		) RETURNING created_at, updated_at;
	`
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	now := time.Now().UTC()
	d.CreatedAt = now
	d.UpdatedAt = now

	return r.pool.QueryRow(
		ctx,
		query,
		d.ID,
		d.OrganizationID,
		d.ServerID,
		d.DomainName,
		d.TargetType,
		d.TargetID,
		d.TargetPort,
		d.Status,
		d.VerificationToken,
		d.SSLStatus,
		d.AutoSSL,
		d.CloudflareDNSManaged,
		d.CloudflareRecordID,
		d.ErrorMessage,
		d.LastCheckedAt,
		d.CreatedAt,
		d.UpdatedAt,
	).Scan(&d.CreatedAt, &d.UpdatedAt)
}

func (r *DomainRepository) GetByID(ctx context.Context, orgID, id uuid.UUID) (*domain.CustomDomain, error) {
	query := `
		SELECT 
			d.id, d.organization_id, d.server_id, COALESCE(s.name, ''), COALESCE(s.public_ip, ''),
			d.domain_name, d.target_type, d.target_id, d.target_port, d.status,
			d.verification_token, d.ssl_status, d.auto_ssl, d.cloudflare_dns_managed,
			COALESCE(d.cloudflare_record_id, ''), COALESCE(d.error_message, ''),
			d.last_checked_at, d.created_at, d.updated_at
		FROM domains d
		LEFT JOIN servers s ON d.server_id = s.id
		WHERE d.id = $1 AND d.organization_id = $2;
	`

	var d domain.CustomDomain
	var serverID *uuid.UUID
	var sName, sIP, cfRecID, errMsg string

	err := r.pool.QueryRow(ctx, query, id, orgID).Scan(
		&d.ID,
		&d.OrganizationID,
		&serverID,
		&sName,
		&sIP,
		&d.DomainName,
		&d.TargetType,
		&d.TargetID,
		&d.TargetPort,
		&d.Status,
		&d.VerificationToken,
		&d.SSLStatus,
		&d.AutoSSL,
		&d.CloudflareDNSManaged,
		&cfRecID,
		&errMsg,
		&d.LastCheckedAt,
		&d.CreatedAt,
		&d.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get domain: %w", err)
	}

	d.ServerID = serverID
	d.ServerName = sName
	d.ServerPublicIP = sIP
	d.CloudflareRecordID = cfRecID
	d.ErrorMessage = errMsg

	return &d, nil
}

func (r *DomainRepository) GetByDomainName(ctx context.Context, domainName string) (*domain.CustomDomain, error) {
	query := `
		SELECT 
			d.id, d.organization_id, d.server_id, COALESCE(s.name, ''), COALESCE(s.public_ip, ''),
			d.domain_name, d.target_type, d.target_id, d.target_port, d.status,
			d.verification_token, d.ssl_status, d.auto_ssl, d.cloudflare_dns_managed,
			COALESCE(d.cloudflare_record_id, ''), COALESCE(d.error_message, ''),
			d.last_checked_at, d.created_at, d.updated_at
		FROM domains d
		LEFT JOIN servers s ON d.server_id = s.id
		WHERE d.domain_name = $1;
	`

	var d domain.CustomDomain
	var serverID *uuid.UUID
	var sName, sIP, cfRecID, errMsg string

	err := r.pool.QueryRow(ctx, query, domainName).Scan(
		&d.ID,
		&d.OrganizationID,
		&serverID,
		&sName,
		&sIP,
		&d.DomainName,
		&d.TargetType,
		&d.TargetID,
		&d.TargetPort,
		&d.Status,
		&d.VerificationToken,
		&d.SSLStatus,
		&d.AutoSSL,
		&d.CloudflareDNSManaged,
		&cfRecID,
		&errMsg,
		&d.LastCheckedAt,
		&d.CreatedAt,
		&d.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get domain by name: %w", err)
	}

	d.ServerID = serverID
	d.ServerName = sName
	d.ServerPublicIP = sIP
	d.CloudflareRecordID = cfRecID
	d.ErrorMessage = errMsg

	return &d, nil
}

func (r *DomainRepository) ListByOrg(ctx context.Context, orgID uuid.UUID) ([]domain.CustomDomain, error) {
	query := `
		SELECT 
			d.id, d.organization_id, d.server_id, COALESCE(s.name, ''), COALESCE(s.public_ip, ''),
			d.domain_name, d.target_type, d.target_id, d.target_port, d.status,
			d.verification_token, d.ssl_status, d.auto_ssl, d.cloudflare_dns_managed,
			COALESCE(d.cloudflare_record_id, ''), COALESCE(d.error_message, ''),
			d.last_checked_at, d.created_at, d.updated_at
		FROM domains d
		LEFT JOIN servers s ON d.server_id = s.id
		WHERE d.organization_id = $1
		ORDER BY d.created_at DESC;
	`

	rows, err := r.pool.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list domains: %w", err)
	}
	defer rows.Close()

	domains := make([]domain.CustomDomain, 0)
	for rows.Next() {
		var d domain.CustomDomain
		var serverID *uuid.UUID
		var sName, sIP, cfRecID, errMsg string

		if err := rows.Scan(
			&d.ID,
			&d.OrganizationID,
			&serverID,
			&sName,
			&sIP,
			&d.DomainName,
			&d.TargetType,
			&d.TargetID,
			&d.TargetPort,
			&d.Status,
			&d.VerificationToken,
			&d.SSLStatus,
			&d.AutoSSL,
			&d.CloudflareDNSManaged,
			&cfRecID,
			&errMsg,
			&d.LastCheckedAt,
			&d.CreatedAt,
			&d.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan domain: %w", err)
		}

		d.ServerID = serverID
		d.ServerName = sName
		d.ServerPublicIP = sIP
		d.CloudflareRecordID = cfRecID
		d.ErrorMessage = errMsg
		domains = append(domains, d)
	}

	return domains, nil
}

func (r *DomainRepository) Update(ctx context.Context, d *domain.CustomDomain) error {
	query := `
		UPDATE domains SET
			server_id = $1,
			target_type = $2,
			target_id = $3,
			target_port = $4,
			status = $5,
			ssl_status = $6,
			auto_ssl = $7,
			cloudflare_dns_managed = $8,
			cloudflare_record_id = $9,
			error_message = $10,
			last_checked_at = $11,
			updated_at = $12
		WHERE id = $13 AND organization_id = $14;
	`
	d.UpdatedAt = time.Now().UTC()
	cmd, err := r.pool.Exec(
		ctx,
		query,
		d.ServerID,
		d.TargetType,
		d.TargetID,
		d.TargetPort,
		d.Status,
		d.SSLStatus,
		d.AutoSSL,
		d.CloudflareDNSManaged,
		d.CloudflareRecordID,
		d.ErrorMessage,
		d.LastCheckedAt,
		d.UpdatedAt,
		d.ID,
		d.OrganizationID,
	)
	if err != nil {
		return fmt.Errorf("failed to update domain: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *DomainRepository) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	query := `DELETE FROM domains WHERE id = $1 AND organization_id = $2;`
	cmd, err := r.pool.Exec(ctx, query, id, orgID)
	if err != nil {
		return fmt.Errorf("failed to delete domain: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *DomainRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.DomainStatus, errMsg string) error {
	query := `
		UPDATE domains SET
			status = $1,
			error_message = $2,
			last_checked_at = $3,
			updated_at = $4
		WHERE id = $5;
	`
	now := time.Now().UTC()
	var errVal sql.NullString
	if errMsg != "" {
		errVal = sql.NullString{String: errMsg, Valid: true}
	}
	_, err := r.pool.Exec(ctx, query, status, errVal, now, now, id)
	return err
}

func (r *DomainRepository) UpdateSSL(ctx context.Context, id uuid.UUID, sslStatus domain.SSLStatus) error {
	query := `
		UPDATE domains SET
			ssl_status = $1,
			updated_at = $2
		WHERE id = $3;
	`
	now := time.Now().UTC()
	_, err := r.pool.Exec(ctx, query, sslStatus, now, id)
	return err
}
