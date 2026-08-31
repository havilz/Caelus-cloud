-- Caelus Cloud - Create Volumes Table (Block Storage Management)
-- Migration: 000010_create_volumes.up.sql

CREATE TABLE IF NOT EXISTS volumes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    server_id UUID REFERENCES servers(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    size_gb INT NOT NULL CHECK (size_gb > 0),
    type VARCHAR(50) NOT NULL DEFAULT 'nvme',
    fs_type VARCHAR(50) NOT NULL DEFAULT 'ext4',
    mount_path VARCHAR(255) NOT NULL DEFAULT '/mnt/data',
    status VARCHAR(50) NOT NULL DEFAULT 'available',
    attached_container_name VARCHAR(255) DEFAULT '',
    iops INT NOT NULL DEFAULT 3000,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_volumes_org_name UNIQUE (organization_id, name)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_volumes_org ON volumes(organization_id);
CREATE INDEX IF NOT EXISTS idx_volumes_status ON volumes(status);
CREATE INDEX IF NOT EXISTS idx_volumes_server ON volumes(server_id);

-- Auto-update updated_at timestamp trigger
DROP TRIGGER IF EXISTS update_volumes_updated_at ON volumes;
CREATE TRIGGER update_volumes_updated_at
    BEFORE UPDATE ON volumes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Row Level Security (RLS) disabled for direct backend query access
ALTER TABLE volumes DISABLE ROW LEVEL SECURITY;
