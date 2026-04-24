package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

var validSeverities = map[string]bool{"minor": true, "major": true, "critical": true}

// subUpdateBody is shared between create/update decoders.
type subUpdateBody struct {
	MinSeverity string  `json:"min_severity"`
	EmailAlerts *bool   `json:"email_alerts"`
	EmailDigest *bool   `json:"email_digest"`
	WebhookURL  *string `json:"webhook_url"`
}

// handleListSubscriptions handles GET /account/subscriptions
func (s *Server) handleListSubscriptions(w http.ResponseWriter, r *http.Request) {
	claims := s.requireAuth(w, r)
	if claims == nil {
		return
	}
	rows, err := s.auth.Store.ListSubscriptionsByUser(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list subscriptions")
		return
	}
	writeEnvelope(w, coalesceSlice(subscriptionResponses(rows)))
}

// handleCreateSubscription handles POST /account/subscriptions
func (s *Server) handleCreateSubscription(w http.ResponseWriter, r *http.Request) {
	claims := s.requireAuth(w, r)
	if claims == nil {
		return
	}
	var body struct {
		ProviderID string `json:"provider_id"`
		subUpdateBody
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ProviderID == "" {
		writeError(w, http.StatusBadRequest, "provider_id required")
		return
	}
	if body.MinSeverity == "" {
		body.MinSeverity = "major"
	}
	if !validSeverities[body.MinSeverity] {
		writeError(w, http.StatusBadRequest, "min_severity must be minor, major or critical")
		return
	}
	emailAlerts, emailDigest := resolveEmailPrefs(body.EmailAlerts, body.EmailDigest)

	sub, err := s.auth.Store.CreateSubscription(r.Context(), pgstore.CreateSubscriptionParams{
		UserID:      claims.UserID,
		ProviderID:  body.ProviderID,
		MinSeverity: body.MinSeverity,
		EmailAlerts: emailAlerts,
		EmailDigest: emailDigest,
		WebhookUrl:  pgtype.Text{String: strOrEmpty(body.WebhookURL), Valid: body.WebhookURL != nil},
	})
	if err != nil {
		code, msg := classifySubCreateError(err)
		writeError(w, code, msg)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": subscriptionResponse(sub, "")})
}

// handleUpdateSubscription handles PUT /account/subscriptions/{id}
func (s *Server) handleUpdateSubscription(w http.ResponseWriter, r *http.Request) {
	claims := s.requireAuth(w, r)
	if claims == nil {
		return
	}
	id, ok := parseID(w, r.PathValue("id"))
	if !ok {
		return
	}
	var body subUpdateBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if body.MinSeverity != "" && !validSeverities[body.MinSeverity] {
		writeError(w, http.StatusBadRequest, "min_severity must be minor, major or critical")
		return
	}

	existing, ok2 := s.fetchOwnedSub(r.Context(), w, id, claims.UserID)
	if !ok2 {
		return
	}

	severity, emailAlerts, emailDigest, webhookURL := mergeSubFields(existing, body)
	sub, err := s.auth.Store.UpdateSubscription(r.Context(), pgstore.UpdateSubscriptionParams{
		ID:          id,
		UserID:      claims.UserID,
		MinSeverity: severity,
		EmailAlerts: emailAlerts,
		EmailDigest: emailDigest,
		WebhookUrl:  webhookURL,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "subscription not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not update subscription")
		return
	}
	writeEnvelope(w, subscriptionResponse(sub, ""))
}

// handleDeleteSubscription handles DELETE /account/subscriptions/{id}
func (s *Server) handleDeleteSubscription(w http.ResponseWriter, r *http.Request) {
	claims := s.requireAuth(w, r)
	if claims == nil {
		return
	}
	id, ok := parseID(w, r.PathValue("id"))
	if !ok {
		return
	}
	if err := s.auth.Store.DeleteSubscription(r.Context(), pgstore.DeleteSubscriptionParams{
		ID:     id,
		UserID: claims.UserID,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "could not delete subscription")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---- helpers ---------------------------------------------------------------

// fetchOwnedSub fetches a subscription and verifies the caller owns it.
func (s *Server) fetchOwnedSub(ctx context.Context, w http.ResponseWriter, id, userID int64) (pgstore.Subscription, bool) {
	existing, err := s.auth.Store.GetSubscription(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "subscription not found")
			return pgstore.Subscription{}, false
		}
		writeError(w, http.StatusInternalServerError, "could not fetch subscription")
		return pgstore.Subscription{}, false
	}
	if existing.UserID != userID {
		writeError(w, http.StatusNotFound, "subscription not found")
		return pgstore.Subscription{}, false
	}
	return existing, true
}

// mergeSubFields overlays non-zero body values onto an existing subscription.
func mergeSubFields(existing pgstore.Subscription, body subUpdateBody) (severity string, emailAlerts bool, emailDigest bool, webhookURL pgtype.Text) {
	severity = existing.MinSeverity
	if body.MinSeverity != "" {
		severity = body.MinSeverity
	}
	emailAlerts = existing.EmailAlerts
	if body.EmailAlerts != nil {
		emailAlerts = *body.EmailAlerts
	}
	emailDigest = existing.EmailDigest
	if body.EmailDigest != nil {
		emailDigest = *body.EmailDigest
	}
	webhookURL = existing.WebhookUrl
	if body.WebhookURL != nil {
		webhookURL = pgtype.Text{String: *body.WebhookURL, Valid: *body.WebhookURL != ""}
	}
	return
}

// resolveEmailPrefs converts optional booleans to concrete values, defaulting to true.
func resolveEmailPrefs(ea, ed *bool) (emailAlerts, emailDigest bool) {
	emailAlerts = ea == nil || *ea
	emailDigest = ed == nil || *ed
	return
}

// classifySubCreateError maps DB constraint violations to HTTP status codes.
func classifySubCreateError(err error) (int, string) {
	if strings.Contains(err.Error(), "subscriptions_user_id_provider_id_key") {
		return http.StatusConflict, "already subscribed to this provider"
	}
	if strings.Contains(err.Error(), "subscriptions_provider_id_fkey") {
		return http.StatusNotFound, "provider not found"
	}
	return http.StatusInternalServerError, "could not create subscription"
}

type subscriptionJSON struct {
	ID           int64   `json:"id"`
	ProviderID   string  `json:"provider_id"`
	ProviderName string  `json:"provider_name,omitempty"`
	MinSeverity  string  `json:"min_severity"`
	EmailAlerts  bool    `json:"email_alerts"`
	EmailDigest  bool    `json:"email_digest"`
	WebhookURL   *string `json:"webhook_url"`
	RSSURL       string  `json:"rss_url"`
}

func subscriptionResponse(s pgstore.Subscription, providerName string) subscriptionJSON {
	return subscriptionJSON{
		ID:           s.ID,
		ProviderID:   s.ProviderID,
		ProviderName: providerName,
		MinSeverity:  s.MinSeverity,
		EmailAlerts:  s.EmailAlerts,
		EmailDigest:  s.EmailDigest,
		WebhookURL:   textVal(s.WebhookUrl),
		RSSURL:       "/v1/providers/" + s.ProviderID + "/feed.xml",
	}
}

func subscriptionResponses(rows []pgstore.ListSubscriptionsByUserRow) []subscriptionJSON {
	out := make([]subscriptionJSON, len(rows))
	for i, r := range rows {
		out[i] = subscriptionResponse(pgstore.Subscription{
			ID:          r.ID,
			UserID:      r.UserID,
			ProviderID:  r.ProviderID,
			MinSeverity: r.MinSeverity,
			EmailAlerts: r.EmailAlerts,
			EmailDigest: r.EmailDigest,
			WebhookUrl:  r.WebhookUrl,
			CreatedAt:   r.CreatedAt,
		}, r.ProviderName)
	}
	return out
}

func parseID(w http.ResponseWriter, s string) (int64, bool) {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}

func strOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
