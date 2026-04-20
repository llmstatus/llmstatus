BEGIN;
DELETE FROM providers WHERE id IN ('moonshot', 'zhipu', 'zeroone_ai', 'qwen', 'fireworks');
COMMIT;
