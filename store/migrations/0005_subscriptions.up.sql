CREATE TABLE subscriptions (
    id           BIGSERIAL   PRIMARY KEY,
    user_id      BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider_id  TEXT        NOT NULL REFERENCES providers(id) ON DELETE CASCADE,
    min_severity TEXT        NOT NULL DEFAULT 'major'
                             CHECK (min_severity IN ('minor', 'major', 'critical')),
    email_alerts BOOLEAN     NOT NULL DEFAULT TRUE,
    email_digest BOOLEAN     NOT NULL DEFAULT TRUE,
    webhook_url  TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, provider_id)
);

CREATE TABLE alert_log (
    id              BIGSERIAL   PRIMARY KEY,
    subscription_id BIGINT      NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    incident_id     UUID        NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    channel         TEXT        NOT NULL CHECK (channel IN ('email', 'webhook')),
    sent_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (subscription_id, incident_id, channel)
);

CREATE INDEX idx_subscriptions_user     ON subscriptions (user_id);
CREATE INDEX idx_subscriptions_provider ON subscriptions (provider_id);
CREATE INDEX idx_alert_log_sub          ON alert_log (subscription_id);
