ALTER TABLE alert_log DROP CONSTRAINT alert_log_unique;
ALTER TABLE alert_log DROP COLUMN event;
ALTER TABLE alert_log ADD CONSTRAINT alert_log_subscription_id_incident_id_channel_key
    UNIQUE (subscription_id, incident_id, channel);
