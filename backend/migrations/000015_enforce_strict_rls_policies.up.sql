-- 000015_enforce_strict_rls_policies.up.sql
-- Phase 7.2 (C-3): True Database Multi-Tenant Row Level Security
--
-- Pendekatan: Setiap transaksi yang berjalan sebagai role "caelus_app" (bukan superuser)
-- WAJIB meng-SET LOCAL app.current_org_id = '<uuid>' sebelum mengakses tabel berikut.
-- Policy memblokir akses secara default (DENY-ALL) jika session variable belum di-set.
--
-- CATATAN: Backend Go menggunakan pgx `BeforeAcquire` hook untuk meng-inject
-- `SET app.current_org_id = ...` di awal setiap transaksi.

-- ============================================================
-- 1. FORCE ROW LEVEL SECURITY pada seluruh tabel multi-tenant
-- ============================================================
ALTER TABLE servers           FORCE ROW LEVEL SECURITY;
ALTER TABLE credentials       FORCE ROW LEVEL SECURITY;
ALTER TABLE providers         FORCE ROW LEVEL SECURITY;
ALTER TABLE organization_members FORCE ROW LEVEL SECURITY;
ALTER TABLE audit_logs        FORCE ROW LEVEL SECURITY;
ALTER TABLE networks          FORCE ROW LEVEL SECURITY;
ALTER TABLE volumes           FORCE ROW LEVEL SECURITY;
ALTER TABLE deployments       FORCE ROW LEVEL SECURITY;
ALTER TABLE deployment_logs   FORCE ROW LEVEL SECURITY;
ALTER TABLE security_scans    FORCE ROW LEVEL SECURITY;
ALTER TABLE security_findings FORCE ROW LEVEL SECURITY;
ALTER TABLE automation_rules  FORCE ROW LEVEL SECURITY;
ALTER TABLE backup_policies   FORCE ROW LEVEL SECURITY;
ALTER TABLE backup_records    FORCE ROW LEVEL SECURITY;
ALTER TABLE iac_configurations FORCE ROW LEVEL SECURITY;

-- ============================================================
-- 2. ENABLE ROW LEVEL SECURITY (jika belum aktif)
-- ============================================================
ALTER TABLE servers           ENABLE ROW LEVEL SECURITY;
ALTER TABLE credentials       ENABLE ROW LEVEL SECURITY;
ALTER TABLE providers         ENABLE ROW LEVEL SECURITY;
ALTER TABLE organization_members ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_logs        ENABLE ROW LEVEL SECURITY;
ALTER TABLE networks          ENABLE ROW LEVEL SECURITY;
ALTER TABLE volumes           ENABLE ROW LEVEL SECURITY;
ALTER TABLE deployments       ENABLE ROW LEVEL SECURITY;
ALTER TABLE deployment_logs   ENABLE ROW LEVEL SECURITY;
ALTER TABLE security_scans    ENABLE ROW LEVEL SECURITY;
ALTER TABLE security_findings ENABLE ROW LEVEL SECURITY;
ALTER TABLE automation_rules  ENABLE ROW LEVEL SECURITY;
ALTER TABLE backup_policies   ENABLE ROW LEVEL SECURITY;
ALTER TABLE backup_records    ENABLE ROW LEVEL SECURITY;
ALTER TABLE iac_configurations ENABLE ROW LEVEL SECURITY;

-- ============================================================
-- 3. DROP POLICY LAMA (idempoten — aman dijalankan ulang)
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
-- 4. BUAT POLICY ISOLASI PER ORGANISASI
--    Prinsip: current_setting('app.current_org_id', true)
--             mengembalikan '' jika belum di-set → DENY ALL
-- ============================================================

-- servers: isolasi berdasarkan organization_id
CREATE POLICY servers_org_isolation ON servers
    AS RESTRICTIVE
    FOR ALL
    USING (
        organization_id::TEXT = current_setting('app.current_org_id', true)
        AND current_setting('app.current_org_id', true) <> ''
    );

-- credentials: isolasi berdasarkan organization_id
CREATE POLICY credentials_org_isolation ON credentials
    AS RESTRICTIVE
    FOR ALL
    USING (
        organization_id::TEXT = current_setting('app.current_org_id', true)
        AND current_setting('app.current_org_id', true) <> ''
    );

-- providers: isolasi berdasarkan organization_id
CREATE POLICY providers_org_isolation ON providers
    AS RESTRICTIVE
    FOR ALL
    USING (
        organization_id::TEXT = current_setting('app.current_org_id', true)
        AND current_setting('app.current_org_id', true) <> ''
    );

-- organization_members: isolasi berdasarkan org_id
CREATE POLICY org_members_isolation ON organization_members
    AS RESTRICTIVE
    FOR ALL
    USING (
        org_id::TEXT = current_setting('app.current_org_id', true)
        AND current_setting('app.current_org_id', true) <> ''
    );

-- audit_logs: isolasi berdasarkan org_id
CREATE POLICY audit_logs_org_isolation ON audit_logs
    AS RESTRICTIVE
    FOR ALL
    USING (
        org_id::TEXT = current_setting('app.current_org_id', true)
        AND current_setting('app.current_org_id', true) <> ''
    );

-- networks: isolasi berdasarkan organization_id
CREATE POLICY networks_org_isolation ON networks
    AS RESTRICTIVE
    FOR ALL
    USING (
        organization_id::TEXT = current_setting('app.current_org_id', true)
        AND current_setting('app.current_org_id', true) <> ''
    );

-- volumes: isolasi berdasarkan organization_id
CREATE POLICY volumes_org_isolation ON volumes
    AS RESTRICTIVE
    FOR ALL
    USING (
        organization_id::TEXT = current_setting('app.current_org_id', true)
        AND current_setting('app.current_org_id', true) <> ''
    );

-- deployments: isolasi berdasarkan organization_id
CREATE POLICY deployments_org_isolation ON deployments
    AS RESTRICTIVE
    FOR ALL
    USING (
        organization_id::TEXT = current_setting('app.current_org_id', true)
        AND current_setting('app.current_org_id', true) <> ''
    );

-- deployment_logs: isolasi melalui JOIN ke deployments (mewarisi org_id parent)
CREATE POLICY deployment_logs_org_isolation ON deployment_logs
    AS RESTRICTIVE
    FOR ALL
    USING (
        EXISTS (
            SELECT 1 FROM deployments d
            WHERE d.id = deployment_logs.deployment_id
              AND d.organization_id::TEXT = current_setting('app.current_org_id', true)
              AND current_setting('app.current_org_id', true) <> ''
        )
    );

-- security_scans: isolasi berdasarkan org_id
CREATE POLICY security_scans_org_isolation ON security_scans
    AS RESTRICTIVE
    FOR ALL
    USING (
        org_id::TEXT = current_setting('app.current_org_id', true)
        AND current_setting('app.current_org_id', true) <> ''
    );

-- security_findings: isolasi melalui JOIN ke security_scans
CREATE POLICY security_findings_org_isolation ON security_findings
    AS RESTRICTIVE
    FOR ALL
    USING (
        EXISTS (
            SELECT 1 FROM security_scans sc
            WHERE sc.id = security_findings.scan_id
              AND sc.org_id::TEXT = current_setting('app.current_org_id', true)
              AND current_setting('app.current_org_id', true) <> ''
        )
    );

-- automation_rules: isolasi berdasarkan org_id
CREATE POLICY automation_rules_org_isolation ON automation_rules
    AS RESTRICTIVE
    FOR ALL
    USING (
        org_id::TEXT = current_setting('app.current_org_id', true)
        AND current_setting('app.current_org_id', true) <> ''
    );

-- backup_policies: isolasi berdasarkan organization_id
CREATE POLICY backup_policies_org_isolation ON backup_policies
    AS RESTRICTIVE
    FOR ALL
    USING (
        organization_id::TEXT = current_setting('app.current_org_id', true)
        AND current_setting('app.current_org_id', true) <> ''
    );

-- backup_records: isolasi berdasarkan organization_id
CREATE POLICY backup_records_org_isolation ON backup_records
    AS RESTRICTIVE
    FOR ALL
    USING (
        organization_id::TEXT = current_setting('app.current_org_id', true)
        AND current_setting('app.current_org_id', true) <> ''
    );

-- iac_configurations: isolasi berdasarkan organization_id
CREATE POLICY iac_configurations_org_isolation ON iac_configurations
    AS RESTRICTIVE
    FOR ALL
    USING (
        organization_id::TEXT = current_setting('app.current_org_id', true)
        AND current_setting('app.current_org_id', true) <> ''
    );
