package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
)

const otpLen = 6

// GenerateOTP returns a random 6-digit code and its SHA-256 hex hash.
func GenerateOTP() (plain, hash string, err error) {
	max := big.NewInt(1_000_000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", "", fmt.Errorf("generate otp: %w", err)
	}
	plain = fmt.Sprintf("%06d", n.Int64())
	hash = VerifyOTPHash(plain)
	return plain, hash, nil
}

// VerifyOTPHash returns the SHA-256 hex hash of a plain code.
// Use this to look up or verify a code without storing the plain text.
func VerifyOTPHash(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return fmt.Sprintf("%x", sum)
}
