package api

import (
	"encoding/json"
	"net/http"
	"time"

	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

// handleUpdateSettings handles PUT /account/settings
// Body: {"digest_hour": 8, "timezone": "Asia/Shanghai"}
func (s *Server) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	claims := s.requireAuth(w, r)
	if claims == nil {
		return
	}
	var body struct {
		DigestHour *int    `json:"digest_hour"`
		Timezone   *string `json:"timezone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	user, err := s.auth.Store.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not fetch user")
		return
	}

	digestHour := int(user.DigestHour)
	if body.DigestHour != nil {
		if *body.DigestHour < 0 || *body.DigestHour > 23 {
			writeError(w, http.StatusBadRequest, "digest_hour must be 0–23")
			return
		}
		digestHour = *body.DigestHour
	}

	timezone := user.Timezone
	if body.Timezone != nil {
		if _, err := time.LoadLocation(*body.Timezone); err != nil {
			writeError(w, http.StatusBadRequest, "invalid timezone")
			return
		}
		timezone = *body.Timezone
	}

	if err := s.auth.Store.UpdateUserSettings(r.Context(), pgstore.UpdateUserSettingsParams{
		ID:         claims.UserID,
		DigestHour: int32(digestHour), //nolint:gosec
		Timezone:   timezone,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "could not update settings")
		return
	}

	writeEnvelope(w, map[string]any{
		"digest_hour": digestHour,
		"timezone":    timezone,
	})
}
