BEGIN;

DELETE FROM sponsors WHERE created_at < '2026-04-26 00:00:00+00';

ALTER TABLE sponsors DROP CONSTRAINT IF EXISTS sponsors_user_or_system;
ALTER TABLE sponsors ALTER COLUMN user_id SET NOT NULL;

ALTER TABLE sponsors DROP COLUMN IF EXISTS is_system;
ALTER TABLE sponsors DROP COLUMN IF EXISTS tagline;

ALTER TABLE sponsors DROP CONSTRAINT IF EXISTS sponsors_tier_check;
ALTER TABLE sponsors ADD CONSTRAINT sponsors_tier_check
    CHECK (tier IN ('founding', 'standard'));

COMMIT;
