CREATE TABLE users (
    id          BIGSERIAL   PRIMARY KEY,
    email       TEXT        NOT NULL UNIQUE,
    digest_hour INT         NOT NULL DEFAULT 8 CHECK (digest_hour BETWEEN 0 AND 23),
    timezone    TEXT        NOT NULL DEFAULT 'UTC',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    verified_at TIMESTAMPTZ
);

CREATE TABLE oauth_accounts (
    id          BIGSERIAL   PRIMARY KEY,
    user_id     BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider    TEXT        NOT NULL CHECK (provider IN ('google', 'github')),
    sub         TEXT        NOT NULL,
    email       TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (provider, sub)
);

CREATE TABLE otp_tokens (
    id          BIGSERIAL   PRIMARY KEY,
    user_id     BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash   TEXT        NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    used_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_otp_tokens_user_expires ON otp_tokens (user_id, expires_at);
