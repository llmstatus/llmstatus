BEGIN;
DELETE FROM providers WHERE id IN ('deepseek', 'google_gemini', 'mistral');
COMMIT;
