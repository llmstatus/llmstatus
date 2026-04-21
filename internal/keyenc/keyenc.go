// Package keyenc encrypts and decrypts short strings (API keys) using
// AES-256-GCM. The master key is a 32-byte value supplied by the caller.
package keyenc

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
)

// Encrypter holds a compiled AES-GCM cipher for a single master key.
type Encrypter struct {
	gcm cipher.AEAD
}

// New creates an Encrypter from a 32-byte hex-encoded master key.
// Typical usage: keyenc.New(os.Getenv("SPONSOR_KEY_ENCRYPTION_KEY"))
func New(hexKey string) (*Encrypter, error) {
	raw, err := hex.DecodeString(hexKey)
	if err != nil || len(raw) != 32 {
		return nil, errors.New("keyenc: SPONSOR_KEY_ENCRYPTION_KEY must be 64 hex chars (32 bytes)")
	}
	block, err := aes.NewCipher(raw)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Encrypter{gcm: gcm}, nil
}

// Encrypt encrypts plaintext and returns a hex-encoded nonce+ciphertext.
func (e *Encrypter) Encrypt(plaintext string) (string, error) {
	nonce := make([]byte, e.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ct := e.gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ct), nil
}

// Decrypt decrypts a hex-encoded nonce+ciphertext produced by Encrypt.
func (e *Encrypter) Decrypt(hexCT string) (string, error) {
	ct, err := hex.DecodeString(hexCT)
	if err != nil {
		return "", err
	}
	ns := e.gcm.NonceSize()
	if len(ct) < ns {
		return "", errors.New("keyenc: ciphertext too short")
	}
	plain, err := e.gcm.Open(nil, ct[:ns], ct[ns:], nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// Hint returns the last 4 characters of the key for display purposes.
func Hint(key string) string {
	if len(key) <= 4 {
		return "****"
	}
	return "..." + key[len(key)-4:]
}
