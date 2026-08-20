-- Migration: 000005_create_storage_and_backups.up.sql
-- Description: Membuat tabel buckets, backup_policies, dan backup_records dengan RLS dan indeks performa tinggi.

-- 1. Tabel Buckets (Metadata Bucket Multi-Provider Tenant)
CREATE TABLE IF NOT EXISTS buckets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(63) NOT NULL UNIQUE,
    provider_type VARCHAR(32) NOT NULL DEFAULT 'minio',
    region VARCHAR(64) NOT NULL DEFAULT 'us-east-1',
    is_public BOOLEAN NOT NULL DEFAULT false,
    versioning BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_buckets_org_id ON buckets(organization_id);
CREATE INDEX IF NOT EXISTS idx_buckets_name ON buckets(name);
CREATE INDEX IF NOT EXISTS idx_buckets_created_at ON buckets(created_at DESC);

-- Trigger updated_at untuk tabel buckets
CREATE TRIGGER update_buckets_modtime
    BEFORE UPDATE ON buckets
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Enable Row Level Security (RLS) pada tabel buckets
ALTER TABLE buckets ENABLE ROW LEVEL SECURITY;

CREATE POLICY buckets_tenant_isolation ON buckets
    FOR ALL
    USING (organization_id = (current_setting('app.current_org_id', true))::uuid)
    WITH CHECK (organization_id = (current_setting('app.current_org_id', true))::uuid);

-- 2. Tabel Backup Policies (Konfigurasi Jadwal & Retensi Backup Server)
CREATE TABLE IF NOT EXISTS backup_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    server_id UUID NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    bucket_id UUID REFERENCES buckets(id) ON DELETE SET NULL,
    name VARCHAR(128) NOT NULL,
    cron_expression VARCHAR(64) NOT NULL DEFAULT '0 2 * * *',
    retention_days INT NOT NULL DEFAULT 7,
    include_disks BOOLEAN NOT NULL DEFAULT true,
    is_active BOOLEAN NOT NULL DEFAULT true,
    next_run_at TIMESTAMPTZ,
    last_run_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_backup_policies_org_id ON backup_policies(organization_id);
CREATE INDEX IF NOT EXISTS idx_backup_policies_server_id ON backup_policies(server_id);
CREATE INDEX IF NOT EXISTS idx_backup_policies_active ON backup_policies(is_active) WHERE is_active = true;

-- Trigger updated_at untuk tabel backup_policies
CREATE TRIGGER update_backup_policies_modtime
    BEFORE UPDATE ON backup_policies
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Enable Row Level Security (RLS) pada tabel backup_policies
ALTER TABLE backup_policies ENABLE ROW LEVEL SECURITY;

CREATE POLICY backup_policies_tenant_isolation ON backup_policies
    FOR ALL
    USING (organization_id = (current_setting('app.current_org_id', true))::uuid)
    WITH CHECK (organization_id = (current_setting('app.current_org_id', true))::uuid);

-- 3. Tabel Backup Records (Riwayat Snapshot & File Arsip Backup)
CREATE TABLE IF NOT EXISTS backup_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    policy_id UUID REFERENCES backup_policies(id) ON DELETE SET NULL,
    server_id UUID NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    bucket_id UUID REFERENCES buckets(id) ON DELETE SET NULL,
    backup_name VARCHAR(255) NOT NULL,
    storage_key VARCHAR(512) NOT NULL,
    size_bytes BIGINT NOT NULL DEFAULT 0,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    error_message TEXT,
    checksum_sha256 VARCHAR(64),
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_backup_records_org_id ON backup_records(organization_id);
CREATE INDEX IF NOT EXISTS idx_backup_records_server_id ON backup_records(server_id);
CREATE INDEX IF NOT EXISTS idx_backup_records_policy_id ON backup_records(policy_id);
CREATE INDEX IF NOT EXISTS idx_backup_records_status ON backup_records(status);
CREATE INDEX IF NOT EXISTS idx_backup_records_expires_at ON backup_records(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_backup_records_started_at ON backup_records(started_at DESC);

-- Trigger updated_at untuk tabel backup_records
CREATE TRIGGER update_backup_records_modtime
    BEFORE UPDATE ON backup_records
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Enable Row Level Security (RLS) pada tabel backup_records
ALTER TABLE backup_records ENABLE ROW LEVEL SECURITY;

CREATE POLICY backup_records_tenant_isolation ON backup_records
    FOR ALL
    USING (organization_id = (current_setting('app.current_org_id', true))::uuid)
    WITH CHECK (organization_id = (current_setting('app.current_org_id', true))::uuid);
