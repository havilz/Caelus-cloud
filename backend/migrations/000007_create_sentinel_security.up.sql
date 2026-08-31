-- Migration: 000007_create_sentinel_security.up.sql
-- Subsystem: Phase 5 Sentinel Security Subsystem

-- 1. Tabel security_scans
CREATE TABLE IF NOT EXISTS security_scans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    server_id UUID REFERENCES servers(id) ON DELETE CASCADE,
    scan_type VARCHAR(50) NOT NULL DEFAULT 'full', -- 'full', 'port', 'tls', 'headers', 'host_config', 'vuln'
    status VARCHAR(50) NOT NULL DEFAULT 'pending',   -- 'pending', 'running', 'completed', 'failed'
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    total_findings INT NOT NULL DEFAULT 0,
    critical_count INT NOT NULL DEFAULT 0,
    high_count INT NOT NULL DEFAULT 0,
    medium_count INT NOT NULL DEFAULT 0,
    low_count INT NOT NULL DEFAULT 0,
    score INT NOT NULL DEFAULT 100, -- 0 to 100
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_security_scans_org ON security_scans(organization_id);
CREATE INDEX IF NOT EXISTS idx_security_scans_server ON security_scans(server_id);
CREATE INDEX IF NOT EXISTS idx_security_scans_status ON security_scans(status);
CREATE INDEX IF NOT EXISTS idx_security_scans_created ON security_scans(created_at DESC);

-- 2. Tabel security_findings
CREATE TABLE IF NOT EXISTS security_findings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    server_id UUID REFERENCES servers(id) ON DELETE CASCADE,
    scan_id UUID REFERENCES security_scans(id) ON DELETE SET NULL,
    fingerprint VARCHAR(255) NOT NULL, -- unik per kombinasi server, category, check_id untuk deduplikasi
    category VARCHAR(50) NOT NULL,    -- 'network', 'tls', 'http_headers', 'host_config', 'vulnerability'
    severity VARCHAR(20) NOT NULL,    -- 'critical', 'high', 'medium', 'low', 'info'
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    evidence JSONB DEFAULT '{}'::jsonb,
    recommendation TEXT,
    remediation_command TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'open', -- 'open', 'acknowledged', 'resolved', 'false_positive'
    resolved_at TIMESTAMPTZ,
    first_detected_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_detected_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_security_findings_org ON security_findings(organization_id);
CREATE INDEX IF NOT EXISTS idx_security_findings_server ON security_findings(server_id);
CREATE INDEX IF NOT EXISTS idx_security_findings_scan ON security_findings(scan_id);
CREATE INDEX IF NOT EXISTS idx_security_findings_severity ON security_findings(severity);
CREATE INDEX IF NOT EXISTS idx_security_findings_status ON security_findings(status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_security_findings_fingerprint ON security_findings(organization_id, server_id, fingerprint) WHERE status != 'resolved';

-- 3. Tabel security_incidents
CREATE TABLE IF NOT EXISTS security_incidents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'high',
    status VARCHAR(50) NOT NULL DEFAULT 'open', -- 'open', 'investigating', 'mitigated', 'closed'
    finding_ids UUID[] DEFAULT '{}',
    summary TEXT,
    mitigation_notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_security_incidents_org ON security_incidents(organization_id);
CREATE INDEX IF NOT EXISTS idx_security_incidents_status ON security_incidents(status);

-- 4. Triggers for updated_at
DROP TRIGGER IF EXISTS update_security_scans_updated_at ON security_scans;
CREATE TRIGGER update_security_scans_updated_at
    BEFORE UPDATE ON security_scans
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_security_incidents_updated_at ON security_incidents;
CREATE TRIGGER update_security_incidents_updated_at
    BEFORE UPDATE ON security_incidents
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 5. Row Level Security (RLS)
ALTER TABLE security_scans ENABLE ROW LEVEL SECURITY;
ALTER TABLE security_findings ENABLE ROW LEVEL SECURITY;
ALTER TABLE security_incidents ENABLE ROW LEVEL SECURITY;

CREATE POLICY security_scans_tenant_isolation ON security_scans
    FOR ALL
    USING (organization_id = NULLIF(current_setting('app.current_organization_id', true), '')::uuid)
    WITH CHECK (organization_id = NULLIF(current_setting('app.current_organization_id', true), '')::uuid);

CREATE POLICY security_findings_tenant_isolation ON security_findings
    FOR ALL
    USING (organization_id = NULLIF(current_setting('app.current_organization_id', true), '')::uuid)
    WITH CHECK (organization_id = NULLIF(current_setting('app.current_organization_id', true), '')::uuid);

CREATE POLICY security_incidents_tenant_isolation ON security_incidents
    FOR ALL
    USING (organization_id = NULLIF(current_setting('app.current_organization_id', true), '')::uuid)
    WITH CHECK (organization_id = NULLIF(current_setting('app.current_organization_id', true), '')::uuid);
