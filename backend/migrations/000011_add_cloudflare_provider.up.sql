-- Migration: 000011_add_cloudflare_provider.up.sql
-- Description: Add Cloudflare provider to providers table

INSERT INTO providers (id, name, slug, is_active, created_at)
VALUES (
    gen_random_uuid(),
    'Cloudflare',
    'cloudflare',
    true,
    CURRENT_TIMESTAMP
)
ON CONFLICT (slug) DO UPDATE
SET is_active = true, name = 'Cloudflare';
