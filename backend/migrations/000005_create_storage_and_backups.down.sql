-- Migration: 000005_create_storage_and_backups.down.sql
-- Description: Rollback tabel buckets, backup_policies, dan backup_records.

DROP TABLE IF EXISTS backup_records CASCADE;
DROP TABLE IF EXISTS backup_policies CASCADE;
DROP TABLE IF EXISTS buckets CASCADE;
