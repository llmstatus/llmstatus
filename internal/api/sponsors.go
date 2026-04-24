package api

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/llmstatus/llmstatus/internal/keyenc"
	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

func nullText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}

var slugRe = regexp.MustCompile(`^[a-z0-9-]{2,40}$`)

// ── Public ────────────────────────────────────────────────────────────────

// GET /v1/sponsors
func (s *Server) listSponsors(w http.ResponseWriter, r *http.Request) {
	rows, err := s.store.ListActiveSponsors(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list sponsors")
		return
	}
	type sponsorOut struct {
		ID         string  `json:"id"`
		Name       string  `json:"name"`
		WebsiteURL *string `json:"website_url,omitempty"`
		LogoURL    *string `json:"logo_url,omitempty"`
		Tier       string  `json:"tier"`
	}
	out := make([]sponsorOut, len(rows))
	for i, sp := range rows {
		var website, logo *string
		if sp.WebsiteUrl.Valid {
			website = &sp.WebsiteUrl.String
		}
		if sp.LogoUrl.Valid {
			logo = &sp.LogoUrl.String
		}
		out[i] = sponsorOut{
			ID:         sp.ID,
			Name:       sp.Name,
			WebsiteURL: website,
			LogoURL:    logo,
			Tier:       sp.Tier,
		}
	}
	writeJSON(w, http.StatusOK, out)
}

// ── Auth-required ─────────────────────────────────────────────────────────

// POST /v1/sponsor/register
// Body: {"id": "soxai", "name": "soxAI", "website_url": "...", "logo_url": "..."}
func (s *Server) registerSponsor(w http.ResponseWriter, r *http.Request) {
	claims := s.requireAuth(w, r)
	if claims == nil {
		return
	}

	var body struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		WebsiteURL string `json:"website_url"`
		LogoURL    string `json:"logo_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	body.ID = strings.ToLower(strings.TrimSpace(body.ID))
	body.Name = strings.TrimSpace(body.Name)
	if !slugRe.MatchString(body.ID) {
		writeError(w, http.StatusBadRequest, "id must be 2-40 lowercase letters, digits, or hyphens")
		return
	}
	if body.Name == "" {
		writeError(w, http.StatusBadRequest, "name required")
		return
	}

	sp, err := s.store.CreateSponsor(r.Context(), pgstore.CreateSponsorParams{
		ID:         body.ID,
		UserID:     claims.UserID,
		Name:       body.Name,
		WebsiteUrl: nullText(body.WebsiteURL),
		LogoUrl:    nullText(body.LogoURL),
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			writeError(w, http.StatusConflict, "sponsor id or user already registered")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not create sponsor")
		return
	}
	writeJSON(w, http.StatusCreated, sp)
}

// GET /v1/sponsor/me
func (s *Server) getSponsorMe(w http.ResponseWriter, r *http.Request) {
	claims := s.requireAuth(w, r)
	if claims == nil {
		return
	}

	sp, err := s.store.GetSponsorByUserID(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusNotFound, "no sponsor profile found")
		return
	}

	keys, err := s.store.ListSponsorKeys(r.Context(), sp.ID)
	if err != nil {
		keys = nil
	}

	type keyOut struct {
		ProviderID     string `json:"provider_id"`
		KeyHint        string `json:"key_hint"`
		Active         bool   `json:"active"`
		LastVerifiedAt any    `json:"last_verified_at"`
		LastError      any    `json:"last_error"`
	}
	keyList := make([]keyOut, len(keys))
	for i, k := range keys {
		var lva, le any
		if k.LastVerifiedAt.Valid {
			lva = k.LastVerifiedAt.Time
		}
		if k.LastError.Valid {
			le = k.LastError.String
		}
		keyList[i] = keyOut{
			ProviderID:     k.ProviderID,
			KeyHint:        k.KeyHint,
			Active:         k.Active,
			LastVerifiedAt: lva,
			LastError:      le,
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"sponsor": sp,
		"keys":    keyList,
	})
}

// PATCH /v1/sponsor/me
// Body: {"name": "...", "website_url": "...", "logo_url": "..."}
func (s *Server) updateSponsorMe(w http.ResponseWriter, r *http.Request) {
	claims := s.requireAuth(w, r)
	if claims == nil {
		return
	}

	sp, err := s.store.GetSponsorByUserID(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusNotFound, "no sponsor profile found")
		return
	}

	var body struct {
		Name       string `json:"name"`
		WebsiteURL string `json:"website_url"`
		LogoURL    string `json:"logo_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if body.Name == "" {
		body.Name = sp.Name
	}

	updated, err := s.store.UpdateSponsor(r.Context(), pgstore.UpdateSponsorParams{
		ID:         sp.ID,
		Name:       body.Name,
		WebsiteUrl: nullText(body.WebsiteURL),
		LogoUrl:    nullText(body.LogoURL),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not update sponsor")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// PUT /v1/sponsor/me/keys/{provider_id}
// Body: {"key": "sk-..."}
func (s *Server) upsertSponsorKey(w http.ResponseWriter, r *http.Request) {
	claims := s.requireAuth(w, r)
	if claims == nil {
		return
	}
	if s.keyEnc == nil {
		writeError(w, http.StatusServiceUnavailable, "key encryption not configured")
		return
	}

	sp, err := s.store.GetSponsorByUserID(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusNotFound, "no sponsor profile found")
		return
	}

	providerID := r.PathValue("provider_id")
	if providerID == "" {
		writeError(w, http.StatusBadRequest, "provider_id required")
		return
	}

	var body struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Key == "" {
		writeError(w, http.StatusBadRequest, "key required")
		return
	}

	enc, err := s.keyEnc.Encrypt(body.Key)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "encryption failed")
		return
	}

	row, err := s.store.UpsertSponsorKey(r.Context(), pgstore.UpsertSponsorKeyParams{
		SponsorID:    sp.ID,
		ProviderID:   providerID,
		EncryptedKey: enc,
		KeyHint:      keyenc.Hint(body.Key),
	})
	if err != nil {
		if strings.Contains(err.Error(), "foreign key") {
			writeError(w, http.StatusBadRequest, "unknown provider_id")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not save key")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"provider_id": row.ProviderID,
		"key_hint":    row.KeyHint,
		"active":      row.Active,
	})
}

// DELETE /v1/sponsor/me/keys/{provider_id}
func (s *Server) deleteSponsorKey(w http.ResponseWriter, r *http.Request) {
	claims := s.requireAuth(w, r)
	if claims == nil {
		return
	}

	sp, err := s.store.GetSponsorByUserID(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusNotFound, "no sponsor profile found")
		return
	}

	providerID := r.PathValue("provider_id")
	if err := s.store.DeleteSponsorKey(r.Context(), pgstore.DeleteSponsorKeyParams{
		SponsorID:  sp.ID,
		ProviderID: providerID,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "could not delete key")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
