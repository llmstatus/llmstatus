-- Add probe_scope to providers so each prober node only runs probes it can reach.
-- Values: 'global' (all nodes), 'intl' (non-CN nodes only), 'cn' (CN nodes only).

ALTER TABLE providers
    ADD COLUMN probe_scope TEXT NOT NULL DEFAULT 'global'
    CHECK (probe_scope IN ('global', 'intl', 'cn'));

-- International providers are blocked from China probe nodes.
UPDATE providers
SET probe_scope = 'intl'
WHERE id IN (
    'ai21', 'anthropic', 'azure_openai', 'cohere', 'fireworks',
    'google', 'google_gemini', 'groq', 'huggingface', 'mistral',
    'openai', 'perplexity', 'together', 'together_ai', 'xai'
);

-- Chinese domestic providers are accessible from all regions (probe_scope stays 'global').
-- deepseek, doubao, minimax, moonshot, qwen, zeroone_ai, zhipu — no change needed.
