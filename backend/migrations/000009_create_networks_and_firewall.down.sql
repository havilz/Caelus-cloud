-- 000009_create_networks_and_firewall.down.sql

DROP POLICY IF EXISTS firewall_rules_tenant_isolation ON firewall_rules;
DROP POLICY IF EXISTS networks_tenant_isolation ON networks;

DROP TRIGGER IF EXISTS update_firewall_rules_updated_at ON firewall_rules;
DROP TRIGGER IF EXISTS update_networks_updated_at ON networks;

DROP TABLE IF EXISTS firewall_rules;
DROP TABLE IF EXISTS networks;
