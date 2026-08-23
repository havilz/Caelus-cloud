-- Migration: 000008_create_iac_and_deployments.down.sql

DROP TRIGGER IF EXISTS update_deployments_updated_at ON deployments;
DROP TRIGGER IF EXISTS update_iac_configurations_updated_at ON iac_configurations;

DROP TABLE IF EXISTS deployment_logs CASCADE;
DROP TABLE IF EXISTS deployments CASCADE;
DROP TABLE IF EXISTS iac_plans CASCADE;
DROP TABLE IF EXISTS iac_states CASCADE;
DROP TABLE IF EXISTS iac_configurations CASCADE;
