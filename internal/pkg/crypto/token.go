package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

type tokenManager struct{}

func NewTokenManager() *tokenManager {
	return &tokenManager{}
}

func (m *tokenManager) Generate(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (m *tokenManager) Hash(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(hash[:])
}