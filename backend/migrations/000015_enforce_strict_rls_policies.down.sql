-- 000015_enforce_strict_rls_policies.down.sql
-- Rollback: Hapus semua RESTRICTIVE policy yang dibuat pada 000015_up
-- dan kembalikan tabel ke state sebelum FORCE RLS.

DROP POLICY IF EXISTS servers_org_isolation           ON servers;
DROP POLICY IF EXISTS credentials_org_isolation       ON credentials;
DROP POLICY IF EXISTS org_members_isolation           ON organization_members;
DROP POLICY IF EXISTS audit_logs_org_isolation        ON audit_logs;
DROP POLICY IF EXISTS networks_org_isolation          ON networks;
DROP POLICY IF EXISTS volumes_org_isolation           ON volumes;
DROP POLICY IF EXISTS deployments_org_isolation       ON deployments;
DROP POLICY IF EXISTS deployment_logs_org_isolation   ON deployment_logs;
DROP POLICY IF EXISTS security_scans_org_isolation    ON security_scans;
DROP POLICY IF EXISTS security_findings_org_isolation ON security_findings;
DROP POLICY IF EXISTS automation_rules_org_isolation  ON automation_rules;
DROP POLICY IF EXISTS backup_policies_org_isolation   ON backup_policies;
DROP POLICY IF EXISTS backup_records_org_isolation    ON backup_records;
DROP POLICY IF EXISTS iac_configurations_org_isolation ON iac_configurations;

-- Kembalikan ke NO FORCE (tidak memaksa RLS untuk superuser/table owner)
ALTER TABLE servers               NO FORCE ROW LEVEL SECURITY;
ALTER TABLE credentials           NO FORCE ROW LEVEL SECURITY;
ALTER TABLE organization_members  NO FORCE ROW LEVEL SECURITY;
ALTER TABLE audit_logs            NO FORCE ROW LEVEL SECURITY;
ALTER TABLE networks              NO FORCE ROW LEVEL SECURITY;
ALTER TABLE volumes               NO FORCE ROW LEVEL SECURITY;
ALTER TABLE deployments           NO FORCE ROW LEVEL SECURITY;
ALTER TABLE deployment_logs       NO FORCE ROW LEVEL SECURITY;
ALTER TABLE security_scans        NO FORCE ROW LEVEL SECURITY;
ALTER TABLE security_findings     NO FORCE ROW LEVEL SECURITY;
ALTER TABLE automation_rules      NO FORCE ROW LEVEL SECURITY;
ALTER TABLE backup_policies       NO FORCE ROW LEVEL SECURITY;
ALTER TABLE backup_records        NO FORCE ROW LEVEL SECURITY;
ALTER TABLE iac_configurations    NO FORCE ROW LEVEL SECURITY;
