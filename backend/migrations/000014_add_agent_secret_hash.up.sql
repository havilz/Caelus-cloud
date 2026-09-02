-- 000014_add_agent_secret_hash.up.sql
-- Menambahkan kolom hash secret agent dan prefix tampilan pada tabel servers.
-- Digunakan oleh middleware RequireAgentAuth untuk memvalidasi token telemetri dari caelus-agent.

ALTER TABLE servers
    ADD COLUMN IF NOT EXISTS agent_secret_hash   VARCHAR(512),
    ADD COLUMN IF NOT EXISTS agent_secret_prefix VARCHAR(16);
