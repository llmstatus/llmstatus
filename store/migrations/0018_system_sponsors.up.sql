-- 0018_system_sponsors.sql — system-level sponsors and tier restructure (LLMS-XXX)
--
-- Changes:
--   1. user_id becomes nullable so system sponsors need no user account.
--   2. is_system flag prevents accidental deletion through normal flows.
--   3. tagline TEXT for sponsor description shown on the sponsors page.
--   4. tier constraint expanded to platinum / gold / silver (legacy values kept).
--   5. Seeds three Platinum system sponsors (Gold/Silver shown as frontend placeholders when empty).

BEGIN;

-- Allow system sponsors to have no backing user account.
ALTER TABLE sponsors ALTER COLUMN user_id DROP NOT NULL;
ALTER TABLE sponsors ADD COLUMN is_system BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE sponsors ADD COLUMN tagline    TEXT;
ALTER TABLE sponsors ADD CONSTRAINT sponsors_user_or_system
    CHECK (is_system = TRUE OR user_id IS NOT NULL);

-- Expand tier vocabulary; keep 'founding'/'standard' for existing rows.
ALTER TABLE sponsors DROP CONSTRAINT IF EXISTS sponsors_tier_check;
ALTER TABLE sponsors ADD CONSTRAINT sponsors_tier_check
    CHECK (tier IN ('platinum', 'gold', 'silver', 'founding', 'standard'));

-- ── Platinum system sponsors ──────────────────────────────────────────────────

INSERT INTO sponsors (id, user_id, name, website_url, logo_url, tagline, tier, active, status, is_system)
VALUES
    (
        'soxai',
        NULL,
        'SoxAI',
        'https://soxai.io',
        'https://s3.llmstatus.io/sponsors/logo-1024-dark.png',
        'Unified access to 40+ AI providers through a single OpenAI-compatible API. Built for teams that demand reliability, flexibility, and control.',
        'platinum',
        TRUE,
        'approved',
        TRUE
    ),
    (
        'onedotnet',
        NULL,
        'OneDotNet',
        'https://onedotnet.com',
        'https://s3.llmstatus.io/sponsors/onedotnet-appicon-dark.png',
        'AI-Powered Global Network Solutions',
        'platinum',
        TRUE,
        'approved',
        TRUE
    ),
    (
        'fastsox',
        NULL,
        'FastSox',
        'https://fastsox.com',
        'https://s3.llmstatus.io/sponsors/app-icon-512.png',
        'Secure VPN with Zero Trust architecture for individuals and enterprises. Military-grade encryption, global network, no-logs policy.',
        'platinum',
        TRUE,
        'approved',
        TRUE
    )
ON CONFLICT (id) DO NOTHING;

COMMIT;
