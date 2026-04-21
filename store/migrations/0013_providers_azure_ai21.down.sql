BEGIN;
DELETE FROM providers WHERE id IN ('azure_openai', 'ai21');
COMMIT;
