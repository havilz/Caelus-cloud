package postgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/havilz/caelus-cloud/backend/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Migrator struct {
	pool *pgxpool.Pool
}

func NewMigrator(pool *pgxpool.Pool) *Migrator {
	return &Migrator{pool: pool}
}

func (m *Migrator) Up(ctx context.Context, migrationsDir string) error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := m.pool.Exec(ctx, createTableSQL); err != nil {
		return fmt.Errorf("gagal membuat tabel schema_migrations: %w", err)
	}

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("gagal membaca direktori migrasi %s: %w", migrationsDir, err)
	}

	var upFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".up.sql") {
			upFiles = append(upFiles, file.Name())
		}
	}
	sort.Strings(upFiles)

	for _, fileName := range upFiles {
		version := strings.TrimSuffix(fileName, ".up.sql")

		var exists bool
		checkSQL := `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`
		if err := m.pool.QueryRow(ctx, checkSQL, version).Scan(&exists); err != nil {
			return fmt.Errorf("gagal memeriksa status migrasi versi %s: %w", version, err)
		}

		if exists {
			continue
		}

		filePath := filepath.Join(migrationsDir, fileName)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("gagal membaca file migrasi %s: %w", filePath, err)
		}

		tx, err := m.pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("gagal memulai transaksi migrasi %s: %w", version, err)
		}

		if _, err := tx.Exec(ctx, string(content)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("gagal mengeksekusi migrasi %s: %w", version, err)
		}

		insertSQL := `INSERT INTO schema_migrations (version, applied_at) VALUES ($1, $2)`
		if _, err := tx.Exec(ctx, insertSQL, version, time.Now()); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("gagal mencatat riwayat migrasi %s: %w", version, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("gagal melakukan commit transaksi migrasi %s: %w", version, err)
		}

		logger.Info("Migrasi basis data berhasil diaplikasikan", "version", version)
	}

	return nil
}
