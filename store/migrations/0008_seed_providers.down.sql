-- Removes seed rows added by 0008_seed_providers.up.sql.
-- Cascades via FK will clean up models rows automatically.

BEGIN;
DELETE FROM providers WHERE id IN ('openai', 'anthropic');
COMMIT;
