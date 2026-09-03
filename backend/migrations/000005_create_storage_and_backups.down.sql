-- Migration: 000005_create_storage_and_backups.down.sql
-- Description: Rollback buckets, backup_policies, and backup_records tables.

DROP TABLE IF EXISTS backup_records CASCADE;
DROP TABLE IF EXISTS backup_policies CASCADE;
DROP TABLE IF EXISTS buckets CASCADE;
