-- 000014_add_agent_secret_hash.up.sql
-- Add agent secret hash and prefix columns to servers table.
-- Used by RequireAgentAuth middleware to validate telemetry tokens from caelus-agent.

ALTER TABLE servers
    ADD COLUMN IF NOT EXISTS agent_secret_hash   VARCHAR(512),
    ADD COLUMN IF NOT EXISTS agent_secret_prefix VARCHAR(16);
