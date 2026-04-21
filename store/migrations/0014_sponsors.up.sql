CREATE TABLE sponsors (
    id          TEXT        PRIMARY KEY,
    user_id     BIGINT      NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT        NOT NULL,
    website_url TEXT,
    logo_url    TEXT,
    tier        TEXT        NOT NULL DEFAULT 'standard'
                            CHECK (tier IN ('founding', 'standard')),
    active      BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE sponsor_keys (
    id               BIGSERIAL   PRIMARY KEY,
    sponsor_id       TEXT        NOT NULL REFERENCES sponsors(id) ON DELETE CASCADE,
    provider_id      TEXT        NOT NULL REFERENCES providers(id),
    encrypted_key    TEXT        NOT NULL,
    key_hint         TEXT        NOT NULL,
    active           BOOLEAN     NOT NULL DEFAULT TRUE,
    last_verified_at TIMESTAMPTZ,
    last_error       TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (sponsor_id, provider_id)
);
