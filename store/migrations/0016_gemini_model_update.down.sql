-- 0016_gemini_model_update.down.sql

BEGIN;

DELETE FROM models WHERE provider_id = 'google_gemini' AND model_id = 'gemini-2.5-flash';

UPDATE models SET active = TRUE
WHERE provider_id = 'google_gemini'
  AND model_id IN ('gemini-2.0-flash', 'gemini-1.5-pro');

COMMIT;
