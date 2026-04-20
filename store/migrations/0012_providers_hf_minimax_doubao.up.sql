-- 0012 — Hugging Face, Minimax, ByteDance (Doubao) (LLMS-058)

BEGIN;

INSERT INTO providers (id, name, category, base_url, auth_type, status_page_url, documentation_url, region, active)
VALUES
    ('huggingface', 'Hugging Face',       'aggregator',       'https://router.huggingface.co',            'bearer', 'https://status.huggingface.co', 'https://huggingface.co/docs/api-inference', 'global', TRUE),
    ('minimax',     'Minimax',            'chinese_official', 'https://api.minimax.chat',                 'bearer', NULL,                            'https://platform.minimaxi.com/document',    'cn',     TRUE),
    ('doubao',      'ByteDance (Doubao)', 'chinese_official', 'https://ark.cn-beijing.volces.com/api/v3', 'bearer', NULL,                            'https://www.volcengine.com/docs/82379',      'cn',     TRUE)
ON CONFLICT (id) DO NOTHING;

INSERT INTO models (provider_id, model_id, display_name, model_type, active)
VALUES
    ('huggingface', 'meta-llama/Llama-3.2-3B-Instruct', 'Llama 3.2 3B',    'chat', TRUE),
    ('huggingface', 'Qwen/Qwen2.5-7B-Instruct',         'Qwen 2.5 7B',     'chat', TRUE),
    ('minimax',     'abab6.5s-chat',                     'ABAB 6.5s',       'chat', TRUE),
    ('minimax',     'abab6.5-chat',                      'ABAB 6.5',        'chat', TRUE),
    ('doubao',      'doubao-lite-4k',                    'Doubao Lite 4k',  'chat', TRUE),
    ('doubao',      'doubao-pro-32k',                    'Doubao Pro 32k',  'chat', TRUE)
ON CONFLICT (provider_id, model_id) DO NOTHING;

COMMIT;
