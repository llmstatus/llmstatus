-- digest_log deduplicates the daily digest: one row per user per local calendar date.
-- sent_date is stored in the user's own timezone, so "2026-04-20 in Asia/Shanghai"
-- is a separate date from "2026-04-20 in UTC".
CREATE TABLE digest_log (
    id        BIGSERIAL   PRIMARY KEY,
    user_id   BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    sent_date DATE        NOT NULL,
    sent_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, sent_date)
);

CREATE INDEX idx_digest_log_user ON digest_log (user_id);
