-- 0013 — Azure OpenAI, AI21 Labs (LLMS-059)

BEGIN;

INSERT INTO providers (id, name, category, base_url, auth_type, status_page_url, documentation_url, region, active)
VALUES
    ('azure_openai', 'Azure OpenAI',  'official', 'https://YOUR_RESOURCE.openai.azure.com', 'api_key_header', 'https://azure.status.microsoft.com', 'https://learn.microsoft.com/azure/ai-services/openai', 'global', TRUE),
    ('ai21',         'AI21 Labs',     'official', 'https://api.ai21.com/studio/v1',         'bearer',  NULL,                                  'https://docs.ai21.com',                               'global', TRUE)
ON CONFLICT (id) DO NOTHING;

INSERT INTO models (provider_id, model_id, display_name, model_type, active)
VALUES
    ('azure_openai', 'gpt-4o-mini', 'GPT-4o mini',  'chat', TRUE),
    ('ai21',         'jamba-mini',  'Jamba Mini',   'chat', TRUE)
ON CONFLICT (provider_id, model_id) DO NOTHING;

COMMIT;
