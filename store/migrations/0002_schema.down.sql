-- 0002_schema.down.sql — reverses 0002_schema.sql
--
-- WARNING: drops all rows in incidents, models, providers.
-- Only run against environments with disposable data.

BEGIN;

DROP TABLE IF EXISTS incidents;
DROP TABLE IF EXISTS models;
DROP TABLE IF EXISTS providers;

COMMIT;
