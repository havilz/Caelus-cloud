-- 000009_create_networks_and_firewall.up.sql

-- 1. Tabel networks (VPC, Bridge, Overlay)
CREATE TABLE IF NOT EXISTS networks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL DEFAULT 'vpc', -- 'vpc', 'bridge', 'overlay'
    cidr VARCHAR(50) NOT NULL DEFAULT '10.20.0.0/16',
    gateway VARCHAR(50) NOT NULL DEFAULT '10.20.0.1',
    region VARCHAR(50) NOT NULL DEFAULT 'ap-southeast-1',
    driver VARCHAR(50) NOT NULL DEFAULT 'bridge',
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- 'active', 'provisioning', 'idle', 'error'
    attached_servers INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_networks_org ON networks(organization_id);
CREATE INDEX IF NOT EXISTS idx_networks_status ON networks(status);

-- 2. Tabel firewall_rules
CREATE TABLE IF NOT EXISTS firewall_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    network_id UUID REFERENCES networks(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    direction VARCHAR(20) NOT NULL DEFAULT 'inbound', -- 'inbound', 'outbound'
    protocol VARCHAR(20) NOT NULL DEFAULT 'tcp', -- 'tcp', 'udp', 'icmp', 'all'
    port_range VARCHAR(100) NOT NULL DEFAULT '80',
    source VARCHAR(100) NOT NULL DEFAULT '0.0.0.0/0',
    action VARCHAR(20) NOT NULL DEFAULT 'allow', -- 'allow', 'deny'
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_firewall_rules_org ON firewall_rules(organization_id);
CREATE INDEX IF NOT EXISTS idx_firewall_rules_network ON firewall_rules(network_id);

-- 3. Triggers for updated_at
DROP TRIGGER IF EXISTS update_networks_updated_at ON networks;
CREATE TRIGGER update_networks_updated_at
    BEFORE UPDATE ON networks
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_firewall_rules_updated_at ON firewall_rules;
CREATE TRIGGER update_firewall_rules_updated_at
    BEFORE UPDATE ON firewall_rules
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 4. Row Level Security (RLS)
ALTER TABLE networks DISABLE ROW LEVEL SECURITY;
ALTER TABLE firewall_rules DISABLE ROW LEVEL SECURITY;
