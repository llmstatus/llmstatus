-- 0009_providers_deepseek_gemini_mistral.sql — seed rows for new adapters (LLMS-055)

BEGIN;

INSERT INTO providers (id, name, category, base_url, auth_type, status_page_url, documentation_url, region, active)
VALUES
    ('deepseek',     'DeepSeek',     'official', 'https://api.deepseek.com',                    'bearer', NULL,                               'https://platform.deepseek.com/docs',           'global', TRUE),
    ('google_gemini','Google Gemini','official', 'https://generativelanguage.googleapis.com',   'api_key_header', NULL,                       'https://ai.google.dev/gemini-api/docs',         'global', TRUE),
    ('mistral',      'Mistral AI',   'official', 'https://api.mistral.ai',                      'bearer', 'https://status.mistral.ai',        'https://docs.mistral.ai',                       'global', TRUE)
ON CONFLICT (id) DO NOTHING;

INSERT INTO models (provider_id, model_id, display_name, model_type, active)
VALUES
    ('deepseek',      'deepseek-chat',        'DeepSeek Chat',         'chat', TRUE),
    ('deepseek',      'deepseek-reasoner',    'DeepSeek Reasoner',     'chat', TRUE),
    ('google_gemini', 'gemini-2.0-flash',     'Gemini 2.0 Flash',      'chat', TRUE),
    ('google_gemini', 'gemini-1.5-pro',       'Gemini 1.5 Pro',        'chat', TRUE),
    ('mistral',       'mistral-small-latest', 'Mistral Small',         'chat', TRUE),
    ('mistral',       'mistral-large-latest', 'Mistral Large',         'chat', TRUE)
ON CONFLICT (provider_id, model_id) DO NOTHING;

COMMIT;
