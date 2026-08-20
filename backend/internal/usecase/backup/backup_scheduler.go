package backup

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

// Scheduler mengelola eksekusi otomatis kebijakan backup server dan pembersihan retensi berkas kedaluwarsa.
type Scheduler struct {
	backupRepo domain.BackupRepository
	usecase    BackupUsecase
	logger     *slog.Logger
	stopChan   chan struct{}
}

// NewScheduler membuat instance baru Scheduler background worker.
func NewScheduler(backupRepo domain.BackupRepository, usecase BackupUsecase, logger *slog.Logger) *Scheduler {
	return &Scheduler{
		backupRepo: backupRepo,
		usecase:    usecase,
		logger:     logger,
		stopChan:   make(chan struct{}),
	}
}

// Start menjalankan loop ticker scheduler secara asinkron dalam goroutine.
func (s *Scheduler) Start(interval time.Duration) {
	if interval <= 0 {
		interval = 60 * time.Second
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		s.logger.Info("Backup Scheduler background worker started", "interval", interval.String())

		for {
			select {
			case <-s.stopChan:
				s.logger.Info("Backup Scheduler background worker stopped")
				return
			case <-ticker.C:
				s.processDueBackups()
				s.cleanExpiredRetention()
			}
		}
	}()
}

// Stop menghentikan proses latar belakang scheduler.
func (s *Scheduler) Stop() {
	close(s.stopChan)
}

// processDueBackups memeriksa dan mengeksekusi kebijakan backup yang telah jatuh tempo.
func (s *Scheduler) processDueBackups() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	now := time.Now().UTC()
	duePolicies, err := s.backupRepo.ListDuePolicies(ctx, now)
	if err != nil {
		s.logger.Error("Failed to query due backup policies", "error", err)
		return
	}

	for _, policy := range duePolicies {
		s.logger.Info("Executing scheduled backup policy", "policy_id", policy.ID, "server_id", policy.ServerID, "policy_name", policy.Name)

		backupName := fmt.Sprintf("auto-%s-%s", policy.Name, now.Format("20060102-150405"))
		record, err := s.usecase.TriggerBackup(ctx, policy.OrganizationID, policy.ServerID, &policy.ID, backupName)
		if err != nil {
			s.logger.Error("Scheduled backup execution failed", "policy_id", policy.ID, "error", err)
			continue
		}

		s.logger.Info("Scheduled backup completed successfully", "record_id", record.ID, "status", record.Status, "size_bytes", record.SizeBytes)

		// Perbarui jadwal eksekusi selanjutnya (contoh: 24 jam ke depan)
		nextRun := now.Add(24 * time.Hour)
		policy.LastRunAt = &now
		policy.NextRunAt = &nextRun
		_ = s.backupRepo.UpdatePolicy(ctx, &policy)
	}
}

// cleanExpiredRetention menghapus berkas arsip backup yang telah melewati batas hari retensi.
func (s *Scheduler) cleanExpiredRetention() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	cleanedCount, err := s.usecase.CleanExpiredBackups(ctx)
	if err != nil {
		s.logger.Error("Failed to clean expired backups", "error", err)
		return
	}

	if cleanedCount > 0 {
		s.logger.Info("Cleaned expired backup retention archives", "deleted_count", cleanedCount)
	}
}
