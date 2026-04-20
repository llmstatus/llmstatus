-- 0008_seed_providers.sql — initial provider and model seed data (LLMS-054)
--
-- Only providers with working adapters and registered livekeys init() are
-- included here. Add rows for each new adapter via a subsequent migration.

BEGIN;

INSERT INTO providers (id, name, category, base_url, auth_type, status_page_url, documentation_url, region, active)
VALUES
    ('openai',    'OpenAI',    'official', 'https://api.openai.com',         'bearer', 'https://status.openai.com',    'https://platform.openai.com/docs',          'global', TRUE),
    ('anthropic', 'Anthropic', 'official', 'https://api.anthropic.com',      'bearer', 'https://status.anthropic.com', 'https://docs.anthropic.com',                 'global', TRUE)
ON CONFLICT (id) DO NOTHING;

INSERT INTO models (provider_id, model_id, display_name, model_type, active)
VALUES
    -- OpenAI
    ('openai', 'gpt-4o-mini',           'GPT-4o mini',          'chat',      TRUE),
    ('openai', 'gpt-4o',                'GPT-4o',               'chat',      TRUE),
    ('openai', 'text-embedding-3-small', 'text-embedding-3-small', 'embedding', TRUE),
    -- Anthropic
    ('anthropic', 'claude-haiku-4-5-20251001', 'Claude Haiku 4.5', 'chat', TRUE),
    ('anthropic', 'claude-sonnet-4-6',         'Claude Sonnet 4.6', 'chat', TRUE)
ON CONFLICT (provider_id, model_id) DO NOTHING;

COMMIT;
