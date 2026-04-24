package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/llmstatus/llmstatus/internal/auth"
	"github.com/llmstatus/llmstatus/internal/email"
	"github.com/llmstatus/llmstatus/internal/otprl"
	pgstore "github.com/llmstatus/llmstatus/internal/store/postgres/gen"
)

const otpTTL = 10 * time.Minute

// AuthStore is the subset of pgstore.Querier used by auth and subscription handlers.
type AuthStore interface {
	UpsertUser(ctx context.Context, e string) (pgstore.User, error)
	GetUserByEmail(ctx context.Context, email string) (pgstore.User, error)
	GetUserByID(ctx context.Context, id int64) (pgstore.User, error)
	MarkUserVerified(ctx context.Context, id int64) error
	CreateOTPToken(ctx context.Context, arg pgstore.CreateOTPTokenParams) (pgstore.OtpToken, error)
	ConsumeOTPToken(ctx context.Context, arg pgstore.ConsumeOTPTokenParams) (pgstore.OtpToken, error)
	UpsertOAuthAccount(ctx context.Context, arg pgstore.UpsertOAuthAccountParams) (pgstore.OauthAccount, error)
	// subscriptions
	ListSubscriptionsByUser(ctx context.Context, userID int64) ([]pgstore.ListSubscriptionsByUserRow, error)
	GetSubscription(ctx context.Context, id int64) (pgstore.Subscription, error)
	CreateSubscription(ctx context.Context, arg pgstore.CreateSubscriptionParams) (pgstore.Subscription, error)
	UpdateSubscription(ctx context.Context, arg pgstore.UpdateSubscriptionParams) (pgstore.Subscription, error)
	DeleteSubscription(ctx context.Context, arg pgstore.DeleteSubscriptionParams) error
	// user settings
	UpdateUserSettings(ctx context.Context, arg pgstore.UpdateUserSettingsParams) error
}

// AuthConfig holds dependencies for auth handlers.
type AuthConfig struct {
	Store          AuthStore
	Email          *email.Client
	JWTSecret      string
	InternalSecret string        // required for /auth/oauth/upsert; shared with Next.js via INTERNAL_SECRET env var
	OTPLimiter     otprl.Limiter // optional; nil disables per-email rate limiting
}

// handleOTPSend handles POST /auth/otp/send
// Body: {"email": "user@example.com"}
// Creates user if not exists, sends 6-digit OTP via email.
func (s *Server) handleOTPSend(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Email == "" {
		writeError(w, http.StatusBadRequest, "email required")
		return
	}

	if s.auth.OTPLimiter != nil {
		ok, retryAfter, err := s.auth.OTPLimiter.Allow(r.Context(), body.Email)
		if err != nil {
			slog.Warn("auth: otp rate-limit check failed", "err", err)
			// fail open — don't block the user on a Redis outage
		} else if !ok {
			w.Header().Set("Retry-After", fmt.Sprintf("%.0f", retryAfter.Seconds()))
			writeError(w, http.StatusTooManyRequests, "too many OTP requests, try again later")
			return
		}
	}

	user, err := s.auth.Store.UpsertUser(r.Context(), body.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create user")
		return
	}

	plain, hash, err := auth.GenerateOTP()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not generate code")
		return
	}

	if _, err := s.auth.Store.CreateOTPToken(r.Context(), pgstore.CreateOTPTokenParams{
		UserID:    user.ID,
		CodeHash:  hash,
		ExpiresAt: mustTimestamptz(time.Now().Add(otpTTL)),
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "could not store code")
		return
	}

	if err := s.auth.Email.Send(r.Context(), email.Message{
		To:      body.Email,
		Subject: fmt.Sprintf("[llmstatus] Your sign-in code: %s", plain),
		Text:    fmt.Sprintf("Your llmstatus.io sign-in code is: %s\n\nExpires in 10 minutes.", plain),
		HTML:    otpEmailHTML(plain),
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "could not send email")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleOTPVerify handles POST /auth/otp/verify
// Body: {"email": "...", "code": "123456"}
// Returns {"token": "<jwt>"} on success.
func (s *Server) handleOTPVerify(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Email == "" || body.Code == "" {
		writeError(w, http.StatusBadRequest, "email and code required")
		return
	}

	// Always perform both DB queries regardless of whether the email exists —
	// early return on missing email leaks user existence via timing difference.
	user, userErr := s.auth.Store.GetUserByEmail(r.Context(), body.Email)
	codeHash := auth.VerifyOTPHash(body.Code)
	_, consumeErr := s.auth.Store.ConsumeOTPToken(r.Context(), pgstore.ConsumeOTPTokenParams{
		UserID:   user.ID, // 0 when user not found; no otp_token row matches, query returns error
		CodeHash: codeHash,
	})
	if userErr != nil || consumeErr != nil {
		writeError(w, http.StatusUnauthorized, "invalid or expired code")
		return
	}

	if err := s.auth.Store.MarkUserVerified(r.Context(), user.ID); err != nil {
		slog.Warn("auth: markUserVerified failed", "user_id", user.ID, "err", err)
	}
	s.writeToken(w, user)
}

// handleOAuthUpsert handles POST /auth/oauth/upsert (internal, called by Next.js only).
// Requires X-Internal-Token header matching AuthConfig.InternalSecret.
// Body: {"provider": "google"|"github", "sub": "...", "email": "..."}
// Returns {"token": "<jwt>"} on success.
func (s *Server) handleOAuthUpsert(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Internal-Token") != s.auth.InternalSecret {
		writeError(w, http.StatusUnauthorized, "forbidden")
		return
	}
	body, err := decodeOAuthBody(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := s.auth.Store.UpsertUser(r.Context(), body.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create user")
		return
	}
	oauthAcc, err := s.auth.Store.UpsertOAuthAccount(r.Context(), pgstore.UpsertOAuthAccountParams{
		UserID:   user.ID,
		Provider: body.Provider,
		Sub:      body.Sub,
		Email:    body.Email,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not link account")
		return
	}
	// Guard against conflict: if the provider+sub was already linked to a
	// different user, the ON CONFLICT clause preserves the original user_id.
	if oauthAcc.UserID != user.ID {
		writeError(w, http.StatusConflict, "oauth account linked to different user")
		return
	}
	_ = s.auth.Store.MarkUserVerified(r.Context(), user.ID)
	s.writeToken(w, user)
}

// handleMe handles GET /auth/me — returns user from Bearer token.
func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	claims := s.requireAuth(w, r)
	if claims == nil {
		return
	}
	user, err := s.auth.Store.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	writeEnvelope(w, map[string]any{
		"id":          user.ID,
		"email":       user.Email,
		"digest_hour": user.DigestHour,
		"timezone":    user.Timezone,
	})
}

// ---- helpers ------------------------------------------------------------

type oauthBody struct {
	Provider string
	Sub      string
	Email    string
}

// decodeOAuthBody parses and validates the /auth/oauth/upsert request body.
func decodeOAuthBody(r *http.Request) (oauthBody, error) {
	var raw struct {
		Provider string `json:"provider"`
		Sub      string `json:"sub"`
		Email    string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil ||
		raw.Provider == "" || raw.Sub == "" || raw.Email == "" {
		return oauthBody{}, errors.New("provider, sub and email required")
	}
	if raw.Provider != "google" && raw.Provider != "github" {
		return oauthBody{}, errors.New("invalid provider")
	}
	return oauthBody{Provider: raw.Provider, Sub: raw.Sub, Email: raw.Email}, nil
}

func (s *Server) writeToken(w http.ResponseWriter, user pgstore.User) {
	token, err := auth.SignJWT(user.ID, user.Email, s.auth.JWTSecret)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not issue token")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

// requireAuth reads the Bearer token from Authorization header or llms_session cookie.
func (s *Server) requireAuth(w http.ResponseWriter, r *http.Request) *auth.Claims {
	var raw string
	if cookie, err := r.Cookie(auth.CookieName); err == nil {
		raw = cookie.Value
	} else if h := r.Header.Get("Authorization"); len(h) > 7 {
		raw = h[7:] // strip "Bearer "
	}
	if raw == "" {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return nil
	}
	claims, err := auth.ParseJWT(raw, s.auth.JWTSecret)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid token")
		return nil
	}
	return claims
}

func otpEmailHTML(code string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html><body style="font-family:monospace;background:#0d0d0d;color:#e0e0e0;padding:40px">
<h2 style="color:#e0e0e0">[llmstatus] sign-in code</h2>
<p style="font-size:32px;letter-spacing:8px;color:#f5a623;font-weight:bold">%s</p>
<p style="color:#888">Expires in 10 minutes. If you didn't request this, ignore it.</p>
</body></html>`, code)
}
