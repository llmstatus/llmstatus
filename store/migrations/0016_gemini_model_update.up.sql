-- 0016_gemini_model_update.sql — switch Google Gemini to gemini-2.5-flash

BEGIN;

-- Deactivate models no longer available on the current API key
UPDATE models SET active = FALSE
WHERE provider_id = 'google_gemini'
  AND model_id IN ('gemini-2.0-flash', 'gemini-1.5-pro');

-- Add gemini-2.5-flash
INSERT INTO models (provider_id, model_id, display_name, model_type, active)
VALUES ('google_gemini', 'gemini-2.5-flash', 'Gemini 2.5 Flash', 'chat', TRUE)
ON CONFLICT (provider_id, model_id) DO UPDATE SET active = TRUE;

COMMIT;
