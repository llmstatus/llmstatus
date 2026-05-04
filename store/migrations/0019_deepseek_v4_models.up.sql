-- 0019_deepseek_v4_models.up.sql — switch DeepSeek probed models to V4 (flash + pro)

BEGIN;

-- Retire the old aliases that are no longer probed.
UPDATE models
SET active = FALSE
WHERE provider_id = 'deepseek'
  AND model_id IN ('deepseek-chat', 'deepseek-reasoner');

-- Add V4 models.
INSERT INTO models (provider_id, model_id, display_name, model_type, active)
VALUES
    ('deepseek', 'deepseek-v4-flash', 'DeepSeek V4 Flash', 'chat', TRUE),
    ('deepseek', 'deepseek-v4-pro',   'DeepSeek V4 Pro',   'chat', TRUE)
ON CONFLICT (provider_id, model_id) DO UPDATE SET active = TRUE;

COMMIT;
