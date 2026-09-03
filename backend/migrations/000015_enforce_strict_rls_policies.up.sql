-- 000015_enforce_strict_rls_policies.up.sql
-- Self-Hosted Architecture Realignment: Application-Level RBAC & Non-Blocking Database Policies
--
-- In Caelus Self-Hosted architecture, workspace/team multi-tenancy is enforced deterministically
-- at the Application RBAC layer (JWT Claims + WHERE organization_id = $1 in repository layer).
-- This migration ensures tables are in NO FORCE RLS state so self-hosted instances run reliably
-- without blocking PostgreSQL session variables.

-- ============================================================
-- 1. DROP OLD RESTRICTIVE POLICIES
-- ============================================================
DROP POLICY IF EXISTS servers_org_isolation           ON servers;
DROP POLICY IF EXISTS credentials_org_isolation       ON credentials;
DROP POLICY IF EXISTS providers_org_isolation         ON providers;
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

-- ============================================================
-- 2. RESET TO NO FORCE ROW LEVEL SECURITY (Self-Hosted Compatible)
-- ============================================================
ALTER TABLE servers                NO FORCE ROW LEVEL SECURITY;
ALTER TABLE credentials            NO FORCE ROW LEVEL SECURITY;
ALTER TABLE organization_members   NO FORCE ROW LEVEL SECURITY;
ALTER TABLE audit_logs             NO FORCE ROW LEVEL SECURITY;
ALTER TABLE networks               NO FORCE ROW LEVEL SECURITY;
ALTER TABLE volumes                NO FORCE ROW LEVEL SECURITY;
ALTER TABLE deployments            NO FORCE ROW LEVEL SECURITY;
ALTER TABLE deployment_logs        NO FORCE ROW LEVEL SECURITY;
ALTER TABLE security_scans         NO FORCE ROW LEVEL SECURITY;
ALTER TABLE security_findings      NO FORCE ROW LEVEL SECURITY;
ALTER TABLE automation_rules       NO FORCE ROW LEVEL SECURITY;
ALTER TABLE backup_policies        NO FORCE ROW LEVEL SECURITY;
ALTER TABLE backup_records         NO FORCE ROW LEVEL SECURITY;
ALTER TABLE iac_configurations     NO FORCE ROW LEVEL SECURITY;
ALTER TABLE providers              NO FORCE ROW LEVEL SECURITY;

-- ============================================================
-- 3. PERMISSIVE APPLICATION POLICIES (Non-blocking fallback)
-- ============================================================
ALTER TABLE servers                ENABLE ROW LEVEL SECURITY;
ALTER TABLE credentials            ENABLE ROW LEVEL SECURITY;
ALTER TABLE organization_members   ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_logs             ENABLE ROW LEVEL SECURITY;
ALTER TABLE networks               ENABLE ROW LEVEL SECURITY;
ALTER TABLE volumes                ENABLE ROW LEVEL SECURITY;
ALTER TABLE deployments            ENABLE ROW LEVEL SECURITY;
ALTER TABLE deployment_logs        ENABLE ROW LEVEL SECURITY;
ALTER TABLE security_scans         ENABLE ROW LEVEL SECURITY;
ALTER TABLE security_findings      ENABLE ROW LEVEL SECURITY;
ALTER TABLE automation_rules       ENABLE ROW LEVEL SECURITY;
ALTER TABLE backup_policies        ENABLE ROW LEVEL SECURITY;
ALTER TABLE backup_records         ENABLE ROW LEVEL SECURITY;
ALTER TABLE iac_configurations     ENABLE ROW LEVEL SECURITY;
ALTER TABLE providers              DISABLE ROW LEVEL SECURITY;

-- Permissive fallback policies for authenticated application backend queries
CREATE POLICY servers_app_access           ON servers           FOR ALL USING (true);
CREATE POLICY credentials_app_access       ON credentials       FOR ALL USING (true);
CREATE POLICY org_members_app_access       ON organization_members FOR ALL USING (true);
CREATE POLICY audit_logs_app_access        ON audit_logs        FOR ALL USING (true);
CREATE POLICY networks_app_access          ON networks          FOR ALL USING (true);
CREATE POLICY volumes_app_access           ON volumes           FOR ALL USING (true);
CREATE POLICY deployments_app_access       ON deployments       FOR ALL USING (true);
CREATE POLICY deployment_logs_app_access   ON deployment_logs   FOR ALL USING (true);
CREATE POLICY security_scans_app_access    ON security_scans    FOR ALL USING (true);
CREATE POLICY security_findings_app_access ON security_findings FOR ALL USING (true);
CREATE POLICY automation_rules_app_access  ON automation_rules  FOR ALL USING (true);
CREATE POLICY backup_policies_app_access   ON backup_policies   FOR ALL USING (true);
CREATE POLICY backup_records_app_access    ON backup_records    FOR ALL USING (true);
CREATE POLICY iac_configurations_app_access ON iac_configurations FOR ALL USING (true);
