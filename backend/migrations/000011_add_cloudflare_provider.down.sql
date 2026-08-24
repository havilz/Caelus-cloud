-- Migration: 000011_add_cloudflare_provider.down.sql
DELETE FROM providers WHERE slug = 'cloudflare';
