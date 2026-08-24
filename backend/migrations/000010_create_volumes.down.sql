-- Caelus Cloud - Drop Volumes Table
-- Migration: 000010_create_volumes.down.sql

DROP TRIGGER IF EXISTS update_volumes_updated_at ON volumes;
DROP TABLE IF EXISTS volumes CASCADE;
