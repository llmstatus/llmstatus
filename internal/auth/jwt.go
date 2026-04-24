// Package auth provides JWT token generation and validation.
package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenTTL is the duration a JWT token is valid.
const TokenTTL = 30 * 24 * time.Hour

// CookieName is the HTTP cookie name for the JWT token.
const CookieName = "llms_session"

// Claims holds the JWT payload fields.
type Claims struct {
	UserID int64  `json:"uid"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// SignJWT issues a signed token valid for 30 days.
func SignJWT(userID int64, email, secret string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("sign jwt: %w", err)
	}
	return signed, nil
}

// ParseJWT validates the token and returns the embedded claims.
func ParseJWT(token, secret string) (*Claims, error) {
	tok, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse jwt: %w", err)
	}
	claims, ok := tok.Claims.(*Claims)
	if !ok || !tok.Valid {
		return nil, fmt.Errorf("invalid jwt claims")
	}
	return claims, nil
}
