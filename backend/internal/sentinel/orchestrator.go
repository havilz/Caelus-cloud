package sentinel

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/internal/sentinel/scanner"
	"github.com/havilz/caelus-cloud/backend/pkg/logger"
)

type ScannerInterface interface {
	Type() domain.ScanType
	Scan(ctx context.Context, target domain.ScanTarget) ([]domain.NormalizedFinding, error)
}

type Orchestrator struct {
	scanners     map[domain.ScanType]ScannerInterface
	normalizer   *FindingNormalizer
	riskEngine   *RiskEngine
	securityRepo domain.SecurityRepository
	eventEmitter func(ctx context.Context, event domain.SystemEvent)
}

func NewOrchestrator(
	securityRepo domain.SecurityRepository,
	eventEmitter func(ctx context.Context, event domain.SystemEvent),
) *Orchestrator {
	o := &Orchestrator{
		scanners:     make(map[domain.ScanType]ScannerInterface),
		normalizer:   NewFindingNormalizer(),
		riskEngine:   NewRiskEngine(),
		securityRepo: securityRepo,
		eventEmitter: eventEmitter,
	}

	o.RegisterScanner(scanner.NewPortScanner(800 * time.Millisecond))
	o.RegisterScanner(scanner.NewTLSScanner(3 * time.Second))
	o.RegisterScanner(scanner.NewHeadersScanner(3 * time.Second))
	o.RegisterScanner(scanner.NewHostConfigScanner())
	o.RegisterScanner(scanner.NewVulnScanner())

	return o
}

func (o *Orchestrator) RegisterScanner(s ScannerInterface) {
	o.scanners[s.Type()] = s
}

func (o *Orchestrator) ExecuteScan(
	ctx context.Context,
	scan *domain.SecurityScan,
	target domain.ScanTarget,
) (*domain.SecurityScan, error) {
	now := time.Now().UTC()
	scan.Status = domain.ScanStatusRunning
	scan.StartedAt = &now
	_ = o.securityRepo.UpdateScan(ctx, scan)

	var targetScanners []ScannerInterface
	if scan.ScanType == domain.ScanTypeFull {
		for _, s := range o.scanners {
			targetScanners = append(targetScanners, s)
		}
	} else if s, exists := o.scanners[scan.ScanType]; exists {
		targetScanners = append(targetScanners, s)
	} else {
		errMsg := fmt.Sprintf("Jenis pemindaian '%s' tidak didukung", scan.ScanType)
		scan.Status = domain.ScanStatusFailed
		scan.ErrorMessage = errMsg
		_ = o.securityRepo.UpdateScan(ctx, scan)
		return scan, fmt.Errorf("%s", errMsg)
	}

	var rawFindings []domain.NormalizedFinding
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, sc := range targetScanners {
		wg.Add(1)
		go func(worker ScannerInterface) {
			defer wg.Done()
			results, err := worker.Scan(ctx, target)
			if err != nil {
				logger.Warn("Scanner failed to execute inspection", "type", worker.Type(), "error", err)
				return
			}
			mu.Lock()
			rawFindings = append(rawFindings, results...)
			mu.Unlock()
		}(sc)
	}

	wg.Wait()

	var normalizedFindings []domain.SecurityFinding
	for _, raw := range rawFindings {
		finding, err := o.normalizer.Normalize(scan.OrganizationID, scan.ServerID, &scan.ID, raw)
		if err != nil {
			continue
		}
		if err := o.securityRepo.UpsertFinding(ctx, finding); err != nil {
			logger.Error("Failed to persist security finding", "title", finding.Title, "error", err)
			continue
		}
		normalizedFindings = append(normalizedFindings, *finding)
	}

	score, critical, high, medium, low := o.riskEngine.CalculateScore(normalizedFindings)
	completedTime := time.Now().UTC()

	scan.Status = domain.ScanStatusCompleted
	scan.CompletedAt = &completedTime
	scan.TotalFindings = len(normalizedFindings)
	scan.CriticalCount = critical
	scan.HighCount = high
	scan.MediumCount = medium
	scan.LowCount = low
	scan.Score = score

	if err := o.securityRepo.UpdateScan(ctx, scan); err != nil {
		return scan, fmt.Errorf("gagal memperbarui hasil scan: %w", err)
	}

	if o.eventEmitter != nil {
		o.eventEmitter(ctx, domain.SystemEvent{
			ID:             uuid.New(),
			OrganizationID: scan.OrganizationID,
			Type:           "security.scan_completed",
			SourceResource: fmt.Sprintf("scan:%s", scan.ID),
			Data: map[string]any{
				"scan_id":        scan.ID.String(),
				"scan_type":      scan.ScanType,
				"score":          score,
				"total_findings": len(normalizedFindings),
				"critical_count": critical,
				"high_count":     high,
			},
			OccurredAt: completedTime,
		})

		if critical > 0 {
			o.eventEmitter(ctx, domain.SystemEvent{
				ID:             uuid.New(),
				OrganizationID: scan.OrganizationID,
				Type:           "security.critical_finding_detected",
				SourceResource: fmt.Sprintf("scan:%s", scan.ID),
				Data: map[string]any{
					"scan_id":        scan.ID.String(),
					"critical_count": critical,
					"server_name":    target.ServerName,
				},
				OccurredAt: completedTime,
			})
		}
	}

	return scan, nil
}
