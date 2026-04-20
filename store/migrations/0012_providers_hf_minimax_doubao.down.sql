BEGIN;
DELETE FROM providers WHERE id IN ('huggingface', 'minimax', 'doubao');
COMMIT;
