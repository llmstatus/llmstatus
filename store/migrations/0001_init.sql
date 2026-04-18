-- 0001_init.sql
--
-- Establishes the directory and numbering convention. Real schema lives
-- in subsequent migrations (see LLMS-002+).

-- Enforce a stable collation so sorting matches what the API returns.
-- Safe no-op on fresh databases; intentionally minimal for LLMS-001.
SELECT 1;
