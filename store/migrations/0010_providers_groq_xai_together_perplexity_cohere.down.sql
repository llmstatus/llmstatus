BEGIN;
DELETE FROM providers WHERE id IN ('groq', 'xai', 'together_ai', 'perplexity', 'cohere');
COMMIT;
