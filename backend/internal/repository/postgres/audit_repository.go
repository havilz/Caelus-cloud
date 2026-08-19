package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

type AuditRepository struct {
	pool *pgxpool.Pool
}

// NewAuditRepository menginisialisasi repository AuditLog berbasis database PostgreSQL.
// Parameter pool merupakan instance koneksi pool *pgxpool.Pool.
// Mengembalikan pointer *AuditRepository yang mengimplementasikan domain.AuditLogRepository.
func NewAuditRepository(pool *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{pool: pool}
}

// Create menyimpan rekaman log audit baru ke dalam tabel audit_logs.
// Parameter ctx merupakan konteks eksekusi query database.
// Parameter log memuat pointer entitas *domain.AuditLog yang akan disimpan.
// Mengembalikan error jika terjadi kegagalan eksekusi query atau serialisasi payload.
func (r *AuditRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	query := `
		INSERT INTO audit_logs (id, organization_id, user_id, action, resource_type, resource_id, ip_address, user_agent, payload, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at;
	`

	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}

	var payloadJSON []byte
	var err error
	if log.Payload != nil {
		payloadJSON, err = json.Marshal(log.Payload)
		if err != nil {
			return fmt.Errorf("gagal melakukan serialisasi payload audit log: %w", err)
		}
	}

	return r.pool.QueryRow(
		ctx,
		query,
		log.ID,
		log.OrganizationID,
		log.UserID,
		log.Action,
		log.ResourceType,
		log.ResourceID,
		log.IPAddress,
		log.UserAgent,
		payloadJSON,
		log.CreatedAt,
	).Scan(&log.ID, &log.CreatedAt)
}

// ListByOrg mengambil daftar log audit berdasarkan organisasi dengan dukungan paginasi.
// Parameter ctx merupakan konteks eksekusi query database.
// Parameter orgID merupakan UUID organisasi yang diaudit.
// Parameter page merupakan nomor halaman data (1-based).
// Parameter limit merupakan jumlah data per halaman.
// Mengembalikan slice []domain.AuditLog, total data int64, dan error jika query gagal.
func (r *AuditRepository) ListByOrg(ctx context.Context, orgID uuid.UUID, page, limit int) ([]domain.AuditLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	countQuery := `SELECT COUNT(*) FROM audit_logs WHERE organization_id = $1;`
	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, orgID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("gagal menghitung total audit logs: %w", err)
	}

	selectQuery := `
		SELECT id, organization_id, user_id, action, resource_type, resource_id, ip_address, user_agent, payload, created_at
		FROM audit_logs
		WHERE organization_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3;
	`

	rows, err := r.pool.Query(ctx, selectQuery, orgID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal mengambil daftar audit logs: %w", err)
	}
	defer rows.Close()

	var logs []domain.AuditLog
	for rows.Next() {
		var l domain.AuditLog
		var payloadJSON []byte
		if err := rows.Scan(
			&l.ID,
			&l.OrganizationID,
			&l.UserID,
			&l.Action,
			&l.ResourceType,
			&l.ResourceID,
			&l.IPAddress,
			&l.UserAgent,
			&payloadJSON,
			&l.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("gagal memindai baris audit log: %w", err)
		}

		if len(payloadJSON) > 0 {
			_ = json.Unmarshal(payloadJSON, &l.Payload)
		}
		logs = append(logs, l)
	}

	return logs, total, nil
}
