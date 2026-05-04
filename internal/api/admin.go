package api

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/llmstatus/llmstatus/internal/email"
)

// requireAdmin validates the Bearer token and verifies the user has is_admin=true.
// Returns true when the caller may proceed; writes an error response and returns false otherwise.
func (s *Server) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	claims := s.requireAuth(w, r)
	if claims == nil {
		return false
	}
	user, err := s.auth.Store.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusForbidden, "forbidden")
		return false
	}
	if !user.IsAdmin {
		writeError(w, http.StatusForbidden, "forbidden")
		return false
	}
	return true
}

// GET /v1/admin/sponsors — list pending sponsors awaiting review.
func (s *Server) adminListSponsors(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	rows, err := s.store.ListPendingSponsors(r.Context())
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
		Status     string  `json:"status"`
		UserID     *int64  `json:"user_id,omitempty"`
		IsSystem   bool    `json:"is_system"`
	}
	out := make([]sponsorOut, len(rows))
	for i, sp := range rows {
		var uid *int64
		if sp.UserID.Valid {
			uid = &sp.UserID.Int64
		}
		out[i] = sponsorOut{
			ID:         sp.ID,
			Name:       sp.Name,
			WebsiteURL: textVal(sp.WebsiteUrl),
			LogoURL:    textVal(sp.LogoUrl),
			Tier:       sp.Tier,
			Status:     sp.Status,
			UserID:     uid,
			IsSystem:   sp.IsSystem,
		}
	}
	writeJSON(w, http.StatusOK, coalesceSlice(out))
}

// POST /v1/admin/sponsors/{id}/approve
func (s *Server) adminApproveSponsor(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	id := r.PathValue("id")
	sp, err := s.store.ApproveSponsor(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "sponsor not found")
		return
	}
	if sp.UserID.Valid {
		s.sendSponsorStatusEmail(r, sp.UserID.Int64, sp.Name, "approved")
	}
	writeJSON(w, http.StatusOK, sp)
}

// POST /v1/admin/sponsors/{id}/reject
func (s *Server) adminRejectSponsor(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	id := r.PathValue("id")
	sp, err := s.store.RejectSponsor(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "sponsor not found")
		return
	}
	if sp.UserID.Valid {
		s.sendSponsorStatusEmail(r, sp.UserID.Int64, sp.Name, "rejected")
	}
	writeJSON(w, http.StatusOK, sp)
}

// POST /v1/admin/test-email — send a test email to verify the Resend integration.
// Requires X-Internal-Token header (shared internal secret). Not restricted to admins
// so it can be called without a user account.
func (s *Server) adminTestEmail(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Internal-Token") != s.auth.InternalSecret {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if s.mailer == nil {
		writeError(w, http.StatusServiceUnavailable, "mailer not configured")
		return
	}
	to := r.URL.Query().Get("to")
	if to == "" {
		writeError(w, http.StatusBadRequest, "to query param required")
		return
	}
	if err := s.mailer.Send(r.Context(), email.Message{
		To:      to,
		Subject: "[llmstatus] Test email — integration check",
		Text:    "If you received this, the Resend integration is working correctly.",
	}); err != nil {
		slog.Error("admin: test email failed", "to", to, "err", err) //nolint:gosec // to is a db-stored email address
		writeError(w, http.StatusInternalServerError, "send failed: "+err.Error())
		return
	}
	slog.Info("admin: test email sent", "to", to) //nolint:gosec // to is a db-stored email address
	writeJSON(w, http.StatusOK, map[string]string{"status": "sent", "to": to})
}

func (s *Server) sendSponsorStatusEmail(r *http.Request, userID int64, sponsorName, status string) {
	if s.mailer == nil || s.auth == nil {
		return
	}
	user, err := s.auth.Store.GetUserByID(r.Context(), userID)
	if err != nil {
		slog.Warn("admin: could not look up sponsor user for email", "user_id", userID, "err", err)
		return
	}
	var subject, body string
	switch status {
	case "approved":
		subject = fmt.Sprintf("[llmstatus] Your sponsor application for %q has been approved", sponsorName)
		body = fmt.Sprintf(
			"Your sponsor profile %q has been approved and is now listed on llmstatus.io/sponsors.\n\n"+
				"You can manage your profile and API keys at https://llmstatus.io/sponsor/dashboard.",
			sponsorName,
		)
	case "rejected":
		subject = fmt.Sprintf("[llmstatus] Your sponsor application for %q was not approved", sponsorName)
		body = fmt.Sprintf(
			"Your sponsor application for %q was reviewed and was not approved at this time.\n\n"+
				"If you have questions, reply to this email.",
			sponsorName,
		)
	}
	if err := s.mailer.Send(r.Context(), email.Message{
		To:      user.Email,
		Subject: subject,
		Text:    body,
	}); err != nil {
		slog.Warn("admin: failed to send sponsor status email", "user_id", userID, "status", status, "err", err)
	}
}
