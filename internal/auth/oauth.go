package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

// GoogleConfig returns OAuth2 config for Google provider.
func GoogleConfig(clientID, clientSecret, redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"openid", "email"},
		Endpoint:     google.Endpoint,
	}
}

// GitHubConfig returns OAuth2 config for GitHub provider.
func GitHubConfig(clientID, clientSecret, redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}
}

// FetchGoogleUser exchanges the code and returns the sub + email from the ID token.
func FetchGoogleUser(ctx context.Context, cfg *oauth2.Config, code string) (sub, email string, err error) {
	tok, err := cfg.Exchange(ctx, code)
	if err != nil {
		return "", "", fmt.Errorf("google exchange: %w", err)
	}
	client := cfg.Client(ctx, tok)
	resp, err := client.Get("https://openidconnect.googleapis.com/v1/userinfo")
	if err != nil {
		return "", "", fmt.Errorf("google userinfo: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck
	var info struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", "", fmt.Errorf("google userinfo decode: %w", err)
	}
	return info.Sub, info.Email, nil
}

// FetchGitHubUser exchanges the code and returns the numeric user ID (as sub) + primary email.
func FetchGitHubUser(ctx context.Context, cfg *oauth2.Config, code string) (sub, email string, err error) {
	tok, err := cfg.Exchange(ctx, code)
	if err != nil {
		return "", "", fmt.Errorf("github exchange: %w", err)
	}
	client := cfg.Client(ctx, tok)

	// Get user id
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("github user: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck
	var user struct {
		ID    int64  `json:"id"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", "", fmt.Errorf("github user decode: %w", err)
	}
	sub = fmt.Sprintf("%d", user.ID)

	if user.Email != "" {
		return sub, user.Email, nil
	}

	// Email not public — fetch from /user/emails
	req2, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	req2.Header.Set("Accept", "application/vnd.github+json")
	resp2, err := client.Do(req2)
	if err != nil {
		return sub, "", fmt.Errorf("github emails: %w", err)
	}
	defer resp2.Body.Close() //nolint:errcheck
	body, _ := io.ReadAll(resp2.Body)
	var emails []struct {
		Email   string `json:"email"`
		Primary bool   `json:"primary"`
	}
	if err := json.Unmarshal(body, &emails); err != nil {
		return sub, "", fmt.Errorf("github emails decode: %w", err)
	}
	for _, e := range emails {
		if e.Primary {
			return sub, e.Email, nil
		}
	}
	return sub, "", fmt.Errorf("github: no primary email found")
}
