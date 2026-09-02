-- Migration: 000012_create_domains_and_ingress.up.sql
-- Description: Create domains table for Custom Domain and Ingress Reverse Proxy routing

CREATE TABLE IF NOT EXISTS domains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    server_id UUID REFERENCES servers(id) ON DELETE SET NULL,
    domain_name VARCHAR(255) NOT NULL,
    target_type VARCHAR(50) NOT NULL DEFAULT 'container',
    target_id VARCHAR(255) NOT NULL,
    target_port INT NOT NULL DEFAULT 80,
    status VARCHAR(50) NOT NULL DEFAULT 'pending_dns',
    verification_token VARCHAR(100) NOT NULL,
    ssl_status VARCHAR(50) NOT NULL DEFAULT 'none',
    auto_ssl BOOLEAN NOT NULL DEFAULT true,
    cloudflare_dns_managed BOOLEAN NOT NULL DEFAULT false,
    cloudflare_record_id VARCHAR(100),
    error_message TEXT,
    last_checked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_org_domain UNIQUE (organization_id, domain_name)
);

CREATE INDEX IF NOT EXISTS idx_domains_org_id ON domains(organization_id);
CREATE INDEX IF NOT EXISTS idx_domains_server_id ON domains(server_id);
CREATE INDEX IF NOT EXISTS idx_domains_domain_name ON domains(domain_name);
CREATE INDEX IF NOT EXISTS idx_domains_status ON domains(status);

-- Enable RLS
ALTER TABLE domains ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS domains_org_isolation ON domains;
CREATE POLICY domains_org_isolation ON domains
    FOR ALL
    USING (organization_id = NULLIF(current_setting('app.current_org_id', true), '')::uuid);

