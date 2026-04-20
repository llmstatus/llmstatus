-- queries.sql — sqlc input for the postgres store (LLMS-005)

-- ============================================================
-- providers
-- ============================================================

-- name: GetProvider :one
SELECT * FROM providers
WHERE id = $1;

-- name: ListProviders :many
SELECT * FROM providers
ORDER BY name;

-- name: ListActiveProviders :many
SELECT * FROM providers
WHERE active = TRUE
ORDER BY name;

-- name: UpsertProvider :exec
INSERT INTO providers (
    id, name, category, base_url, auth_type,
    status_page_url, documentation_url, region, active, config
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
ON CONFLICT (id) DO UPDATE SET
    name              = EXCLUDED.name,
    category          = EXCLUDED.category,
    base_url          = EXCLUDED.base_url,
    auth_type         = EXCLUDED.auth_type,
    status_page_url   = EXCLUDED.status_page_url,
    documentation_url = EXCLUDED.documentation_url,
    region            = EXCLUDED.region,
    active            = EXCLUDED.active,
    config            = EXCLUDED.config;

-- name: SetProviderActive :exec
UPDATE providers SET active = $2 WHERE id = $1;

-- ============================================================
-- models
-- ============================================================

-- name: GetModel :one
SELECT * FROM models
WHERE provider_id = $1 AND model_id = $2;

-- name: ListModelsByProvider :many
SELECT * FROM models
WHERE provider_id = $1
ORDER BY model_id;

-- name: UpsertModel :one
INSERT INTO models (
    provider_id, model_id, display_name, model_type, active
) VALUES (
    $1, $2, $3, $4, $5
)
ON CONFLICT (provider_id, model_id) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    model_type   = EXCLUDED.model_type,
    active       = EXCLUDED.active
RETURNING *;

-- ============================================================
-- incidents
-- ============================================================

-- name: CreateIncident :one
INSERT INTO incidents (
    slug, provider_id, severity, title, description,
    status, affected_models, affected_regions,
    started_at, detection_method, detection_rule, metrics_snapshot
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8,
    $9, $10, $11, $12
)
RETURNING *;

-- name: GetIncidentByID :one
SELECT * FROM incidents WHERE id = $1;

-- name: GetIncidentBySlug :one
SELECT * FROM incidents WHERE slug = $1;

-- name: GetOngoingByProviderAndRule :one
SELECT * FROM incidents
WHERE provider_id    = $1
  AND detection_rule = $2
  AND status         = 'ongoing'
LIMIT 1;

-- name: ListIncidents :many
SELECT * FROM incidents
ORDER BY started_at DESC
LIMIT $1 OFFSET $2;

-- name: ListIncidentsByStatus :many
SELECT * FROM incidents
WHERE status = $1
ORDER BY started_at DESC
LIMIT $2 OFFSET $3;

-- name: ListIncidentsByProvider :many
SELECT * FROM incidents
WHERE provider_id = $1
ORDER BY started_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateIncidentStatus :exec
UPDATE incidents
SET status     = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: ResolveIncident :exec
UPDATE incidents
SET status      = 'resolved',
    resolved_at = $2,
    updated_at  = NOW()
WHERE id = $1;

-- name: SetIncidentDescription :exec
UPDATE incidents
SET description    = $2,
    human_reviewed = $3,
    updated_at     = NOW()
WHERE id = $1;

-- ============================================================
-- auth (LLMS-049)
-- ============================================================

-- name: UpsertUser :one
INSERT INTO users (email)
VALUES ($1)
ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUserSettings :exec
UPDATE users
SET digest_hour = $2,
    timezone    = $3
WHERE id = $1;

-- name: MarkUserVerified :exec
UPDATE users SET verified_at = NOW() WHERE id = $1 AND verified_at IS NULL;

-- name: CreateOTPToken :one
INSERT INTO otp_tokens (user_id, code_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ConsumeOTPToken :one
-- Finds a valid (unused, unexpired) token for the user and marks it used.
UPDATE otp_tokens
SET used_at = NOW()
WHERE id = (
    SELECT t.id FROM otp_tokens t
    WHERE t.user_id   = $1
      AND t.code_hash = $2
      AND t.used_at   IS NULL
      AND t.expires_at > NOW()
    ORDER BY t.created_at DESC
    LIMIT 1
)
RETURNING *;

-- name: UpsertOAuthAccount :one
INSERT INTO oauth_accounts (user_id, provider, sub, email)
VALUES ($1, $2, $3, $4)
ON CONFLICT (provider, sub) DO UPDATE SET email = EXCLUDED.email
RETURNING *;

-- name: GetUserByOAuth :one
SELECT u.* FROM users u
JOIN oauth_accounts o ON o.user_id = u.id
WHERE o.provider = $1 AND o.sub = $2;
