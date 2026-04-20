ALTER TABLE alert_log
    ADD COLUMN event TEXT NOT NULL DEFAULT 'incident.created'
        CHECK (event IN ('incident.created', 'incident.resolved'));

ALTER TABLE alert_log
    DROP CONSTRAINT alert_log_subscription_id_incident_id_channel_key;

ALTER TABLE alert_log
    ADD CONSTRAINT alert_log_unique
        UNIQUE (subscription_id, incident_id, channel, event);
