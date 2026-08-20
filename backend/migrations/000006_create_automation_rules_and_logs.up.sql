-- Migration 000006: Create Automation Rules and Execution Logs Table with RLS
-- Tabel untuk menyimpan aturan otomasi berbasis Event-Condition-Action (ECA)

CREATE TABLE IF NOT EXISTS automation_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    trigger_type VARCHAR(64) NOT NULL, -- metric_threshold, server_status_changed, backup_event, scheduled_cron
    trigger_config JSONB NOT NULL DEFAULT '{}'::jsonb,
    conditions JSONB NOT NULL DEFAULT '[]'::jsonb,
    actions JSONB NOT NULL DEFAULT '[]'::jsonb,
    cooldown_seconds INT NOT NULL DEFAULT 300,
    last_triggered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indeks performa untuk query aturan otomasi berdasarkan organisasi dan status aktif
CREATE INDEX IF NOT EXISTS idx_automation_rules_org ON automation_rules(organization_id);
CREATE INDEX IF NOT EXISTS idx_automation_rules_active ON automation_rules(organization_id, is_active);
CREATE INDEX IF NOT EXISTS idx_automation_rules_trigger ON automation_rules(trigger_type);

-- Trigger untuk memperbarui kolom updated_at secara otomatis
DROP TRIGGER IF EXISTS update_automation_rules_updated_at ON automation_rules;
CREATE TRIGGER update_automation_rules_updated_at
    BEFORE UPDATE ON automation_rules
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Tabel untuk mencatat log riwayat audit eksekusi aturan otomasi
CREATE TABLE IF NOT EXISTS automation_execution_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id UUID NOT NULL REFERENCES automation_rules(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    trigger_event VARCHAR(64) NOT NULL,
    status VARCHAR(32) NOT NULL, -- success, failed, partially_failed, skipped
    evaluated_conditions JSONB NOT NULL DEFAULT '{}'::jsonb,
    executed_actions JSONB NOT NULL DEFAULT '[]'::jsonb,
    error_message TEXT,
    execution_duration_ms INT NOT NULL DEFAULT 0,
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indeks performa untuk riwayat eksekusi
CREATE INDEX IF NOT EXISTS idx_automation_logs_org_rule ON automation_execution_logs(organization_id, rule_id);
CREATE INDEX IF NOT EXISTS idx_automation_logs_executed_at ON automation_execution_logs(organization_id, executed_at DESC);
CREATE INDEX IF NOT EXISTS idx_automation_logs_status ON automation_execution_logs(organization_id, status);

-- Mengaktifkan Row Level Security (RLS) untuk isolasi multi-tenant
ALTER TABLE automation_rules ENABLE ROW LEVEL SECURITY;
ALTER TABLE automation_execution_logs ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS automation_rules_org_isolation ON automation_rules;
CREATE POLICY automation_rules_org_isolation ON automation_rules
    FOR ALL
    USING (
        organization_id IN (
            SELECT organization_id FROM organization_members
            WHERE user_id = auth.uid()
        )
    );

DROP POLICY IF EXISTS automation_logs_org_isolation ON automation_execution_logs;
CREATE POLICY automation_logs_org_isolation ON automation_execution_logs
    FOR ALL
    USING (
        organization_id IN (
            SELECT organization_id FROM organization_members
            WHERE user_id = auth.uid()
        )
    );
