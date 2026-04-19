-- 0002_schema.sql — relational schema: providers, models, incidents (LLMS-005)
--
-- probe samples are stored in InfluxDB 3, not here.
-- This migration runs in a single transaction.

BEGIN;

CREATE TABLE providers (
    id                TEXT        PRIMARY KEY,
    name              TEXT        NOT NULL,
    category          TEXT        NOT NULL,
    base_url          TEXT        NOT NULL,
    auth_type         TEXT        NOT NULL,
    status_page_url   TEXT,
    documentation_url TEXT,
    region            TEXT        NOT NULL DEFAULT 'global',
    added_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    active            BOOLEAN     NOT NULL DEFAULT TRUE,
    config            JSONB       NOT NULL DEFAULT '{}'::JSONB,

    CONSTRAINT providers_category_chk
        CHECK (category IN ('official', 'aggregator', 'chinese_official')),
    CONSTRAINT providers_auth_type_chk
        CHECK (auth_type IN ('bearer', 'api_key_header', 'custom')),
    CONSTRAINT providers_region_chk
        CHECK (region IN ('global', 'us', 'cn', 'eu'))
);

CREATE TABLE models (
    id           BIGSERIAL   PRIMARY KEY,
    provider_id  TEXT        NOT NULL REFERENCES providers(id),
    model_id     TEXT        NOT NULL,
    display_name TEXT        NOT NULL,
    model_type   TEXT        NOT NULL,
    active       BOOLEAN     NOT NULL DEFAULT TRUE,

    CONSTRAINT models_model_type_chk
        CHECK (model_type IN ('chat', 'embedding', 'image')),
    CONSTRAINT models_provider_model_uniq
        UNIQUE (provider_id, model_id)
);

CREATE INDEX idx_models_provider ON models (provider_id);

CREATE TABLE incidents (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    slug             TEXT        UNIQUE NOT NULL,
    provider_id      TEXT        NOT NULL REFERENCES providers(id),
    severity         TEXT        NOT NULL,
    title            TEXT        NOT NULL,
    description      TEXT,
    status           TEXT        NOT NULL,
    affected_models  TEXT[]      NOT NULL DEFAULT '{}',
    affected_regions TEXT[]      NOT NULL DEFAULT '{}',
    started_at       TIMESTAMPTZ NOT NULL,
    resolved_at      TIMESTAMPTZ,
    detection_method TEXT        NOT NULL,
    detection_rule   TEXT,
    metrics_snapshot JSONB       NOT NULL DEFAULT '{}'::JSONB,
    human_reviewed   BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT incidents_severity_chk
        CHECK (severity IN ('minor', 'major', 'critical')),
    CONSTRAINT incidents_status_chk
        CHECK (status IN ('ongoing', 'monitoring', 'resolved')),
    CONSTRAINT incidents_detection_method_chk
        CHECK (detection_method IN ('auto', 'manual'))
);

CREATE INDEX idx_incidents_provider_status ON incidents (provider_id, status);
CREATE INDEX idx_incidents_started_at      ON incidents (started_at DESC);

COMMIT;
