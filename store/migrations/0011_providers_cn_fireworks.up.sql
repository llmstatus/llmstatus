-- 0011 — Moonshot, Zhipu AI, 01.AI, Qwen, Fireworks AI (LLMS-057)
-- Chinese providers use region='cn'; Fireworks uses 'global'.

BEGIN;

INSERT INTO providers (id, name, category, base_url, auth_type, status_page_url, documentation_url, region, active)
VALUES
    ('moonshot',    'Moonshot AI (Kimi)', 'chinese_official', 'https://api.moonshot.cn',                               'bearer', NULL, 'https://platform.moonshot.cn/docs',                    'cn', TRUE),
    ('zhipu',       'Zhipu AI (GLM)',     'chinese_official', 'https://open.bigmodel.cn',                              'bearer', NULL, 'https://bigmodel.cn/dev/api',                          'cn', TRUE),
    ('zeroone_ai',  '01.AI (Yi)',         'chinese_official', 'https://api.01.ai',                                     'bearer', NULL, 'https://platform.01.ai/docs',                         'cn', TRUE),
    ('qwen',        'Qwen (Alibaba)',      'chinese_official', 'https://dashscope.aliyuncs.com/compatible-mode',        'bearer', NULL, 'https://help.aliyun.com/zh/dashscope/developer-reference/api-details', 'cn', TRUE),
    ('fireworks',   'Fireworks AI',       'aggregator',       'https://api.fireworks.ai',                              'bearer', 'https://status.fireworks.ai', 'https://readme.fireworks.ai', 'global', TRUE)
ON CONFLICT (id) DO NOTHING;

INSERT INTO models (provider_id, model_id, display_name, model_type, active)
VALUES
    ('moonshot',   'moonshot-v1-8k',                                        'Moonshot v1 8k',      'chat', TRUE),
    ('moonshot',   'moonshot-v1-32k',                                       'Moonshot v1 32k',     'chat', TRUE),
    ('zhipu',      'glm-4-flash',                                           'GLM-4 Flash',         'chat', TRUE),
    ('zhipu',      'glm-4',                                                 'GLM-4',               'chat', TRUE),
    ('zeroone_ai', 'yi-lightning',                                          'Yi Lightning',        'chat', TRUE),
    ('zeroone_ai', 'yi-large',                                              'Yi Large',            'chat', TRUE),
    ('qwen',       'qwen-turbo',                                            'Qwen Turbo',          'chat', TRUE),
    ('qwen',       'qwen-plus',                                             'Qwen Plus',           'chat', TRUE),
    ('fireworks',  'accounts/fireworks/models/llama-v3p1-8b-instruct',      'Llama 3.1 8B',        'chat', TRUE),
    ('fireworks',  'accounts/fireworks/models/mixtral-8x7b-instruct',       'Mixtral 8x7B',        'chat', TRUE)
ON CONFLICT (provider_id, model_id) DO NOTHING;

COMMIT;
