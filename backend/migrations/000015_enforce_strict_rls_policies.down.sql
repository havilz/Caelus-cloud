-- 000015_enforce_strict_rls_policies.down.sql
-- Rollback: Remove policies created in 000015_up

DROP POLICY IF EXISTS servers_app_access           ON servers;
DROP POLICY IF EXISTS credentials_app_access       ON credentials;
DROP POLICY IF EXISTS org_members_app_access       ON organization_members;
DROP POLICY IF EXISTS audit_logs_app_access        ON audit_logs;
DROP POLICY IF EXISTS networks_app_access          ON networks;
DROP POLICY IF EXISTS volumes_app_access           ON volumes;
DROP POLICY IF EXISTS deployments_app_access       ON deployments;
DROP POLICY IF EXISTS deployment_logs_app_access   ON deployment_logs;
DROP POLICY IF EXISTS security_scans_app_access    ON security_scans;
DROP POLICY IF EXISTS security_findings_app_access ON security_findings;
DROP POLICY IF EXISTS automation_rules_app_access  ON automation_rules;
DROP POLICY IF EXISTS backup_policies_app_access   ON backup_policies;
DROP POLICY IF EXISTS backup_records_app_access    ON backup_records;
DROP POLICY IF EXISTS iac_configurations_app_access ON iac_configurations;

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
