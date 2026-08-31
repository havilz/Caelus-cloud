-- Migration: 000008_create_iac_and_deployments.up.sql
-- Subsystem: Phase 6 Declarative IaC & Container Orchestration

-- 1. Tabel iac_configurations
CREATE TABLE IF NOT EXISTS iac_configurations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    raw_yaml TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'draft', -- 'draft', 'planned', 'applied', 'failed', 'drifted'
    current_version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_iac_configs_org ON iac_configurations(organization_id);
CREATE INDEX IF NOT EXISTS idx_iac_configs_status ON iac_configurations(status);

-- 2. Tabel iac_states
CREATE TABLE IF NOT EXISTS iac_states (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    configuration_id UUID NOT NULL REFERENCES iac_configurations(id) ON DELETE CASCADE,
    version INT NOT NULL,
    state_data JSONB NOT NULL DEFAULT '{}'::jsonb,
    hash VARCHAR(64) NOT NULL,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    applied_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_iac_states_config ON iac_states(configuration_id);
CREATE INDEX IF NOT EXISTS idx_iac_states_version ON iac_states(configuration_id, version DESC);

-- 3. Tabel iac_plans
CREATE TABLE IF NOT EXISTS iac_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    configuration_id UUID NOT NULL REFERENCES iac_configurations(id) ON DELETE CASCADE,
    target_version INT NOT NULL,
    changes JSONB NOT NULL DEFAULT '[]'::jsonb,
    summary JSONB NOT NULL DEFAULT '{}'::jsonb,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'applying', 'applied', 'failed', 'discarded', 'rolled_back'
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    executed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_iac_plans_config ON iac_plans(configuration_id);
CREATE INDEX IF NOT EXISTS idx_iac_plans_status ON iac_plans(status);

-- 4. Tabel deployments (Container Orchestration)
CREATE TABLE IF NOT EXISTS deployments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    server_id UUID REFERENCES servers(id) ON DELETE SET NULL,
    app_name VARCHAR(255) NOT NULL,
    image_tag VARCHAR(255) NOT NULL,
    container_name VARCHAR(255) NOT NULL,
    environment_variables JSONB DEFAULT '{}'::jsonb,
    port_bindings JSONB DEFAULT '[]'::jsonb,
    volume_bindings JSONB DEFAULT '[]'::jsonb,
    restart_policy VARCHAR(50) NOT NULL DEFAULT 'unless-stopped',
    status VARCHAR(50) NOT NULL DEFAULT 'queued', -- 'queued', 'pulling', 'building', 'deploying', 'running', 'failed', 'stopped', 'rolled_back'
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    finished_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_deployments_org ON deployments(organization_id);
CREATE INDEX IF NOT EXISTS idx_deployments_server ON deployments(server_id);
CREATE INDEX IF NOT EXISTS idx_deployments_status ON deployments(status);

-- 5. Tabel deployment_logs
CREATE TABLE IF NOT EXISTS deployment_logs (
    id BIGSERIAL PRIMARY KEY,
    deployment_id UUID NOT NULL REFERENCES deployments(id) ON DELETE CASCADE,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    stream VARCHAR(20) NOT NULL DEFAULT 'stdout', -- 'stdout', 'stderr', 'system'
    message TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_deployment_logs_deployment ON deployment_logs(deployment_id, timestamp ASC);

-- 6. Triggers for updated_at
DROP TRIGGER IF EXISTS update_iac_configurations_updated_at ON iac_configurations;
CREATE TRIGGER update_iac_configurations_updated_at
    BEFORE UPDATE ON iac_configurations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_deployments_updated_at ON deployments;
CREATE TRIGGER update_deployments_updated_at
    BEFORE UPDATE ON deployments
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 7. Row Level Security (RLS)
ALTER TABLE iac_configurations ENABLE ROW LEVEL SECURITY;
ALTER TABLE iac_states ENABLE ROW LEVEL SECURITY;
ALTER TABLE iac_plans ENABLE ROW LEVEL SECURITY;
ALTER TABLE deployments ENABLE ROW LEVEL SECURITY;
ALTER TABLE deployment_logs ENABLE ROW LEVEL SECURITY;

CREATE POLICY iac_configurations_tenant_isolation ON iac_configurations
    FOR ALL
    USING (organization_id = NULLIF(current_setting('app.current_organization_id', true), '')::uuid)
    WITH CHECK (organization_id = NULLIF(current_setting('app.current_organization_id', true), '')::uuid);

CREATE POLICY iac_states_tenant_isolation ON iac_states
    FOR ALL
    USING (
        configuration_id IN (
            SELECT id FROM iac_configurations
            WHERE organization_id = NULLIF(current_setting('app.current_organization_id', true), '')::uuid
        )
    )
    WITH CHECK (
        configuration_id IN (
            SELECT id FROM iac_configurations
            WHERE organization_id = NULLIF(current_setting('app.current_organization_id', true), '')::uuid
        )
    );

CREATE POLICY iac_plans_tenant_isolation ON iac_plans
    FOR ALL
    USING (
        configuration_id IN (
            SELECT id FROM iac_configurations
            WHERE organization_id = NULLIF(current_setting('app.current_organization_id', true), '')::uuid
        )
    )
    WITH CHECK (
        configuration_id IN (
            SELECT id FROM iac_configurations
            WHERE organization_id = NULLIF(current_setting('app.current_organization_id', true), '')::uuid
        )
    );

CREATE POLICY deployments_tenant_isolation ON deployments
    FOR ALL
    USING (organization_id = NULLIF(current_setting('app.current_organization_id', true), '')::uuid)
    WITH CHECK (organization_id = NULLIF(current_setting('app.current_organization_id', true), '')::uuid);

CREATE POLICY deployment_logs_tenant_isolation ON deployment_logs
    FOR ALL
    USING (
        deployment_id IN (
            SELECT id FROM deployments
            WHERE organization_id = NULLIF(current_setting('app.current_organization_id', true), '')::uuid
        )
    )
    WITH CHECK (
        deployment_id IN (
            SELECT id FROM deployments
            WHERE organization_id = NULLIF(current_setting('app.current_organization_id', true), '')::uuid
        )
    );

