-- 000014_add_agent_secret_hash.down.sql
ALTER TABLE servers
    DROP COLUMN IF EXISTS agent_secret_hash,
    DROP COLUMN IF EXISTS agent_secret_prefix;
