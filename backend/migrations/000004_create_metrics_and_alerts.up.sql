-- Caelus Cloud - Metrics & Alerting Tables (Up)

-- 1. Server Metrics Table (Time-Series)
CREATE TABLE IF NOT EXISTS server_metrics (
    id BIGSERIAL PRIMARY KEY,
    server_id UUID NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    cpu_usage_pct NUMERIC(5,2) NOT NULL DEFAULT 0.0,
    memory_used_mb BIGINT NOT NULL DEFAULT 0,
    memory_total_mb BIGINT NOT NULL DEFAULT 0,
    memory_usage_pct NUMERIC(5,2) NOT NULL DEFAULT 0.0,
    disk_used_gb NUMERIC(6,2) NOT NULL DEFAULT 0.0,
    disk_total_gb NUMERIC(6,2) NOT NULL DEFAULT 0.0,
    disk_usage_pct NUMERIC(5,2) NOT NULL DEFAULT 0.0,
    network_in_kb BIGINT NOT NULL DEFAULT 0,
    network_out_kb BIGINT NOT NULL DEFAULT 0,
    network_in_rate_kbps NUMERIC(10,2) NOT NULL DEFAULT 0.0,
    network_out_rate_kbps NUMERIC(10,2) NOT NULL DEFAULT 0.0,
    uptime_seconds BIGINT NOT NULL DEFAULT 0,
    containers_count INT NOT NULL DEFAULT 0,
    docker_available BOOLEAN NOT NULL DEFAULT FALSE,
    containers_json JSONB DEFAULT '[]'::jsonb,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_server_metrics_server_recorded ON server_metrics(server_id, recorded_at DESC);

-- 2. Alert Rules Table
CREATE TABLE IF NOT EXISTS alert_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    server_id UUID REFERENCES servers(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    metric_name VARCHAR(50) NOT NULL,
    operator VARCHAR(10) NOT NULL DEFAULT '>',
    threshold NUMERIC(5,2) NOT NULL,
    duration_seconds INT NOT NULL DEFAULT 60,
    severity VARCHAR(20) NOT NULL DEFAULT 'warning',
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alert_rules_org ON alert_rules(organization_id);
CREATE INDEX IF NOT EXISTS idx_alert_rules_server ON alert_rules(server_id);

CREATE TRIGGER update_alert_rules_updated_at
    BEFORE UPDATE ON alert_rules
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 3. Alerts Table
CREATE TABLE IF NOT EXISTS alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    server_id UUID NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    rule_id UUID REFERENCES alert_rules(id) ON DELETE SET NULL,
    alert_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'warning',
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'active',
    current_value NUMERIC(10,2),
    threshold_value NUMERIC(10,2),
    acknowledged_at TIMESTAMPTZ,
    acknowledged_by UUID REFERENCES users(id) ON DELETE SET NULL,
    resolved_at TIMESTAMPTZ,
    resolved_by UUID REFERENCES users(id) ON DELETE SET NULL,
    triggered_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alerts_org_status ON alerts(organization_id, status);
CREATE INDEX IF NOT EXISTS idx_alerts_server_status ON alerts(server_id, status);
CREATE INDEX IF NOT EXISTS idx_alerts_triggered_at ON alerts(triggered_at DESC);

-- Enable Row Level Security (RLS)
ALTER TABLE server_metrics ENABLE ROW LEVEL SECURITY;
ALTER TABLE alert_rules ENABLE ROW LEVEL SECURITY;
ALTER TABLE alerts ENABLE ROW LEVEL SECURITY;
