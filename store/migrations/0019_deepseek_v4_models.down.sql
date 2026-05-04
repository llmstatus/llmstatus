-- 0019_deepseek_v4_models.down.sql

BEGIN;

DELETE FROM models
WHERE provider_id = 'deepseek'
  AND model_id IN ('deepseek-v4-flash', 'deepseek-v4-pro');

UPDATE models
SET active = TRUE
WHERE provider_id = 'deepseek'
  AND model_id IN ('deepseek-chat', 'deepseek-reasoner');

COMMIT;
