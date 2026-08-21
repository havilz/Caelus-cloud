-- Migration: 000007_create_sentinel_security.down.sql

DROP POLICY IF EXISTS security_incidents_tenant_isolation ON security_incidents;
DROP POLICY IF EXISTS security_findings_tenant_isolation ON security_findings;
DROP POLICY IF EXISTS security_scans_tenant_isolation ON security_scans;

DROP TRIGGER IF EXISTS update_security_incidents_updated_at ON security_incidents;
DROP TRIGGER IF EXISTS update_security_scans_updated_at ON security_scans;

DROP TABLE IF EXISTS security_incidents;
DROP TABLE IF EXISTS security_findings;
DROP TABLE IF EXISTS security_scans;
