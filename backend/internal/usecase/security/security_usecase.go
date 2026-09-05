package security

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/sentinel"
	"github.com/havilz/caelus-cloud/backend/pkg/logger"
)

type SecurityUsecase interface {
	TriggerScan(ctx context.Context, orgID uuid.UUID, serverID *uuid.UUID, scanType domain.ScanType) (*domain.SecurityScan, error)
	GetScan(ctx context.Context, orgID, scanID uuid.UUID) (*domain.SecurityScan, error)
	ListScans(ctx context.Context, orgID uuid.UUID, serverID *uuid.UUID, page, limit int) ([]domain.SecurityScan, int, error)
	ListFindings(ctx context.Context, orgID uuid.UUID, serverID *uuid.UUID, category *domain.FindingCategory, severity *domain.FindingSeverity, status *domain.FindingStatus, page, limit int) ([]domain.SecurityFinding, int, error)
	GetFinding(ctx context.Context, orgID, findingID uuid.UUID) (*domain.SecurityFinding, error)
	UpdateFindingStatus(ctx context.Context, orgID, findingID uuid.UUID, status domain.FindingStatus) error
	GetPostureOverview(ctx context.Context, orgID uuid.UUID) (*domain.SecurityPostureOverview, error)
	CreateIncident(ctx context.Context, orgID uuid.UUID, title, summary string, severity domain.FindingSeverity, findingIDs []uuid.UUID) (*domain.SecurityIncident, error)
	ListIncidents(ctx context.Context, orgID uuid.UUID, status *domain.IncidentStatus, page, limit int) ([]domain.SecurityIncident, int, error)
	UpdateIncidentStatus(ctx context.Context, orgID, incidentID uuid.UUID, status domain.IncidentStatus, notes string) error
}

type securityUsecase struct {
	securityRepo domain.SecurityRepository
	serverRepo   domain.ServerRepository
	metricRepo   domain.MetricRepository
	orchestrator *sentinel.Orchestrator
}

func NewSecurityUsecase(
	securityRepo domain.SecurityRepository,
	serverRepo domain.ServerRepository,
	metricRepo domain.MetricRepository,
	orchestrator *sentinel.Orchestrator,
) SecurityUsecase {
	return &securityUsecase{
		securityRepo: securityRepo,
		serverRepo:   serverRepo,
		metricRepo:   metricRepo,
		orchestrator: orchestrator,
	}
}

func (u *securityUsecase) TriggerScan(ctx context.Context, orgID uuid.UUID, serverID *uuid.UUID, scanType domain.ScanType) (*domain.SecurityScan, error) {
	if scanType == "" {
		scanType = domain.ScanTypeFull
	}

	target := domain.ScanTarget{
		OrganizationID: orgID,
		ServerID:       serverID,
	}

	if serverID != nil && *serverID != uuid.Nil {
		srv, err := u.serverRepo.GetByID(ctx, *serverID)
		if err != nil {
			return nil, fmt.Errorf("server tidak ditemukan: %w", err)
		}
		if srv.OrganizationID != orgID {
			return nil, domain.ErrForbidden
		}

		target.ServerName = srv.Name
		if srv.IPAddress != nil {
			target.IPAddress = *srv.IPAddress
		}
		if srv.Hostname != nil {
			target.Hostname = *srv.Hostname
		}
		target.OSType = srv.OSType

		latestMetric, err := u.metricRepo.GetLatestByServerID(ctx, srv.ID)
		if err == nil && latestMetric != nil {
			target.TelemetryData = &domain.HostMetricsPayload{
				CPUUsagePct:    latestMetric.CPUUsagePct,
				MemoryUsagePct: latestMetric.MemoryUsagePct,
				MemoryUsedMB:   uint64(latestMetric.MemoryUsedMB),
				MemoryTotalMB:  uint64(latestMetric.MemoryTotalMB),
				DiskUsagePct:   latestMetric.DiskUsagePct,
				DiskUsedGB:     latestMetric.DiskUsedGB,
				DiskTotalGB:    latestMetric.DiskTotalGB,
				UptimeSeconds:  uint64(latestMetric.UptimeSeconds),
				Platform:       srv.OSType,
				Hostname:       target.Hostname,
			}
		}
	} else {
		target.ServerName = "All Organization Servers"
		target.IPAddress = "127.0.0.1"
	}

	scan := &domain.SecurityScan{
		ID:             uuid.New(),
		OrganizationID: orgID,
		ServerID:       serverID,
		ServerName:     target.ServerName,
		ScanType:       scanType,
		Status:         domain.ScanStatusPending,
		Score:          100,
	}

	if err := u.securityRepo.CreateScan(ctx, scan); err != nil {
		return nil, fmt.Errorf("gagal membuat sesi scan: %w", err)
	}

	go func(s *domain.SecurityScan, t domain.ScanTarget) {
		bgCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		if _, err := u.orchestrator.ExecuteScan(bgCtx, s, t); err != nil {
			logger.Error("Sentinel scan execution failed", "scan_id", s.ID, "error", err)
		}
	}(scan, target)

	return scan, nil
}

func (u *securityUsecase) GetScan(ctx context.Context, orgID, scanID uuid.UUID) (*domain.SecurityScan, error) {
	return u.securityRepo.GetScanByID(ctx, orgID, scanID)
}

func (u *securityUsecase) ListScans(ctx context.Context, orgID uuid.UUID, serverID *uuid.UUID, page, limit int) ([]domain.SecurityScan, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return u.securityRepo.ListScans(ctx, orgID, serverID, page, limit)
}

func (u *securityUsecase) ListFindings(
	ctx context.Context,
	orgID uuid.UUID,
	serverID *uuid.UUID,
	category *domain.FindingCategory,
	severity *domain.FindingSeverity,
	status *domain.FindingStatus,
	page, limit int,
) ([]domain.SecurityFinding, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return u.securityRepo.ListFindings(ctx, orgID, serverID, category, severity, status, page, limit)
}

func (u *securityUsecase) GetFinding(ctx context.Context, orgID, findingID uuid.UUID) (*domain.SecurityFinding, error) {
	return u.securityRepo.GetFindingByID(ctx, orgID, findingID)
}

func (u *securityUsecase) UpdateFindingStatus(ctx context.Context, orgID, findingID uuid.UUID, status domain.FindingStatus) error {
	return u.securityRepo.UpdateFindingStatus(ctx, orgID, findingID, status)
}

func (u *securityUsecase) GetPostureOverview(ctx context.Context, orgID uuid.UUID) (*domain.SecurityPostureOverview, error) {
	return u.securityRepo.GetPostureOverview(ctx, orgID)
}

func (u *securityUsecase) CreateIncident(ctx context.Context, orgID uuid.UUID, title, summary string, severity domain.FindingSeverity, findingIDs []uuid.UUID) (*domain.SecurityIncident, error) {
	if severity == "" {
		severity = domain.SeverityHigh
	}
	incident := &domain.SecurityIncident{
		OrganizationID: orgID,
		Title:          title,
		Summary:        summary,
		Severity:       severity,
		Status:         domain.IncidentStatusOpen,
		FindingIDs:     findingIDs,
	}
	if err := u.securityRepo.CreateIncident(ctx, incident); err != nil {
		return nil, err
	}
	return incident, nil
}

func (u *securityUsecase) ListIncidents(ctx context.Context, orgID uuid.UUID, status *domain.IncidentStatus, page, limit int) ([]domain.SecurityIncident, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return u.securityRepo.ListIncidents(ctx, orgID, status, page, limit)
}

func (u *securityUsecase) UpdateIncidentStatus(ctx context.Context, orgID, incidentID uuid.UUID, status domain.IncidentStatus, notes string) error {
	return u.securityRepo.UpdateIncidentStatus(ctx, orgID, incidentID, status, notes)
}
