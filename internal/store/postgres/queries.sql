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

-- name: ListProvidersForScope :many
-- Returns active providers visible to a probe node with the given scope.
-- 'global' providers are probed by every node; 'intl'/'cn' are scope-specific.
SELECT * FROM providers
WHERE active = TRUE
  AND (probe_scope = 'global' OR probe_scope = $1)
ORDER BY name;

-- name: UpsertProvider :exec
INSERT INTO providers (
    id, name, category, base_url, auth_type,
    status_page_url, documentation_url, region, active, config, probe_scope
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
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
    config            = EXCLUDED.config,
    probe_scope       = EXCLUDED.probe_scope;

-- name: SetProviderProbeScope :exec
UPDATE providers SET probe_scope = $2 WHERE id = $1;

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
-- user_reports (LLMS-048)
-- ============================================================

-- name: InsertUserReport :exec
-- Inserts a report only when no report from the same ip_hash exists for the
-- same provider within the last 5 minutes (server-side dedup).
INSERT INTO user_reports (provider_id, ip_hash)
SELECT $1, $2
WHERE NOT EXISTS (
    SELECT 1 FROM user_reports
    WHERE provider_id = $1
      AND ip_hash     = $2
      AND created_at  > NOW() - INTERVAL '5 minutes'
);

-- name: UserReportHistogram :many
-- Returns 24 hourly buckets for the given provider, oldest first.
-- Buckets with zero reports are included via generate_series.
SELECT
    gs.bucket,
    COALESCE(COUNT(r.id), 0)::BIGINT AS count
FROM generate_series(
    date_trunc('hour', NOW() - INTERVAL '23 hours'),
    date_trunc('hour', NOW()),
    INTERVAL '1 hour'
) AS gs(bucket)
LEFT JOIN user_reports r
    ON r.provider_id = $1
    AND date_trunc('hour', r.created_at) = gs.bucket
GROUP BY gs.bucket
ORDER BY gs.bucket;

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

-- ============================================================
-- subscriptions (LLMS-050)
-- ============================================================

-- name: ListSubscriptionsByUser :many
SELECT s.*, p.name AS provider_name
FROM subscriptions s
JOIN providers p ON p.id = s.provider_id
WHERE s.user_id = $1
ORDER BY p.name;

-- name: GetSubscription :one
SELECT * FROM subscriptions WHERE id = $1;

-- name: CreateSubscription :one
INSERT INTO subscriptions (user_id, provider_id, min_severity, email_alerts, email_digest, webhook_url)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateSubscription :one
UPDATE subscriptions
SET min_severity = $3,
    email_alerts = $4,
    email_digest = $5,
    webhook_url  = $6
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteSubscription :exec
DELETE FROM subscriptions WHERE id = $1 AND user_id = $2;

-- name: ListIncidentsUpdatedSince :many
SELECT * FROM incidents
WHERE updated_at > $1
ORDER BY updated_at;

-- name: ListSubscriptionsForProvider :many
SELECT s.*, u.email AS user_email, p.name AS provider_name
FROM subscriptions s
JOIN users u ON u.id = s.user_id
JOIN providers p ON p.id = s.provider_id
WHERE s.provider_id = $1
ORDER BY s.id;

-- name: IsAlertSent :one
SELECT EXISTS(
    SELECT 1 FROM alert_log
    WHERE subscription_id = $1
      AND incident_id      = $2
      AND channel          = $3
      AND event            = $4
) AS sent;

-- name: LogAlert :exec
INSERT INTO alert_log (subscription_id, incident_id, channel, event)
VALUES ($1, $2, $3, $4)
ON CONFLICT (subscription_id, incident_id, channel, event) DO NOTHING;

-- ============================================================
-- digest (LLMS-052)
-- ============================================================

-- name: ListUsersForDigest :many
-- Returns distinct users who have at least one email_digest=true subscription.
SELECT DISTINCT u.*
FROM users u
JOIN subscriptions s ON s.user_id = u.id
WHERE s.email_digest = TRUE;

-- name: ListDigestSubscriptions :many
-- Returns all digest-enabled subscriptions for a user with provider data.
SELECT s.*, p.name AS provider_name
FROM subscriptions s
JOIN providers p ON p.id = s.provider_id
WHERE s.user_id = $1 AND s.email_digest = TRUE
ORDER BY p.name;

-- name: ListRecentIncidentsByProvider :many
-- Incidents opened or updated within the last 24 hours for a provider.
SELECT * FROM incidents
WHERE provider_id = $1
  AND updated_at  > NOW() - INTERVAL '24 hours'
ORDER BY started_at DESC;

-- name: IsDigestSent :one
SELECT EXISTS(
    SELECT 1 FROM digest_log
    WHERE user_id   = $1
      AND sent_date = $2
) AS sent;

-- name: LogDigest :exec
INSERT INTO digest_log (user_id, sent_date) VALUES ($1, $2)
ON CONFLICT (user_id, sent_date) DO NOTHING;

-- ── Sponsors ──────────────────────────────────────────────────────────────

-- name: CreateSponsor :one
INSERT INTO sponsors (id, user_id, name, website_url, logo_url, tier)
VALUES ($1, $2, $3, $4, $5, 'standard')
RETURNING *;

-- name: GetSponsorByUserID :one
SELECT * FROM sponsors WHERE user_id = $1;

-- name: GetSponsorByID :one
SELECT * FROM sponsors WHERE id = $1;

-- name: UpdateSponsor :one
UPDATE sponsors
SET name = $2, website_url = $3, logo_url = $4
WHERE id = $1
RETURNING *;

-- name: ListActiveSponsors :many
SELECT * FROM sponsors WHERE active = TRUE ORDER BY tier DESC, created_at ASC;

-- ── Sponsor keys ──────────────────────────────────────────────────────────

-- name: UpsertSponsorKey :one
INSERT INTO sponsor_keys (sponsor_id, provider_id, encrypted_key, key_hint)
VALUES ($1, $2, $3, $4)
ON CONFLICT (sponsor_id, provider_id) DO UPDATE
    SET encrypted_key    = EXCLUDED.encrypted_key,
        key_hint         = EXCLUDED.key_hint,
        active           = TRUE,
        last_error       = NULL,
        updated_at       = NOW()
RETURNING *;

-- name: DeleteSponsorKey :exec
DELETE FROM sponsor_keys WHERE sponsor_id = $1 AND provider_id = $2;

-- name: ListSponsorKeys :many
SELECT * FROM sponsor_keys WHERE sponsor_id = $1 ORDER BY provider_id;

-- name: ListActiveSponsorKeys :many
SELECT * FROM sponsor_keys WHERE active = TRUE ORDER BY sponsor_id, provider_id;

-- name: UpdateSponsorKeyVerification :exec
UPDATE sponsor_keys
SET last_verified_at = NOW(),
    last_error       = $3,
    active           = ($3 = ''),
    updated_at       = NOW()
WHERE sponsor_id = $1 AND provider_id = $2;
