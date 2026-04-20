-- 0010 — seed rows for Groq, xAI, Together AI, Perplexity, Cohere (LLMS-056)

BEGIN;

INSERT INTO providers (id, name, category, base_url, auth_type, status_page_url, documentation_url, region, active)
VALUES
    ('groq',        'Groq',         'official',   'https://api.groq.com',              'bearer', 'https://status.groq.com',          'https://console.groq.com/docs',       'global', TRUE),
    ('xai',         'xAI',          'official',   'https://api.x.ai',                  'bearer', NULL,                               'https://docs.x.ai',                    'global', TRUE),
    ('together_ai', 'Together AI',  'aggregator', 'https://api.together.xyz',          'bearer', 'https://status.together.ai',       'https://docs.together.ai',             'global', TRUE),
    ('perplexity',  'Perplexity',   'official',   'https://api.perplexity.ai',         'bearer', NULL,                               'https://docs.perplexity.ai',           'global', TRUE),
    ('cohere',      'Cohere',       'official',   'https://api.cohere.com',            'bearer', 'https://status.cohere.com',        'https://docs.cohere.com',              'global', TRUE)
ON CONFLICT (id) DO NOTHING;

INSERT INTO models (provider_id, model_id, display_name, model_type, active)
VALUES
    ('groq',        'llama-3.3-70b-versatile',       'Llama 3.3 70B',              'chat', TRUE),
    ('groq',        'gemma2-9b-it',                   'Gemma 2 9B',                 'chat', TRUE),
    ('xai',         'grok-3-mini',                    'Grok 3 Mini',                'chat', TRUE),
    ('xai',         'grok-3',                         'Grok 3',                     'chat', TRUE),
    ('together_ai', 'meta-llama/Llama-3-8b-chat-hf',  'Llama 3 8B',                 'chat', TRUE),
    ('together_ai', 'mistralai/Mixtral-8x7B-Instruct-v0.1', 'Mixtral 8x7B',         'chat', TRUE),
    ('perplexity',  'sonar',                          'Sonar',                      'chat', TRUE),
    ('perplexity',  'sonar-pro',                      'Sonar Pro',                  'chat', TRUE),
    ('cohere',      'command-r',                      'Command R',                  'chat', TRUE),
    ('cohere',      'command-r-plus',                 'Command R+',                 'chat', TRUE)
ON CONFLICT (provider_id, model_id) DO NOTHING;

COMMIT;
