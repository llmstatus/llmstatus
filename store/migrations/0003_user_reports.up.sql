-- 0003_user_reports.sql — anonymous user report counts per provider (LLMS-048)
--
-- ip_hash stores SHA-256(client_ip) — no raw IPs ever written.
-- The 5-minute dedup is enforced at the application layer via NOT EXISTS.

BEGIN;

CREATE TABLE user_reports (
    id          BIGSERIAL   PRIMARY KEY,
    provider_id TEXT        NOT NULL REFERENCES providers(id),
    ip_hash     TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_reports_provider_time ON user_reports (provider_id, created_at DESC);

COMMIT;
