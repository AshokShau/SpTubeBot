package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"
)

var (
	urlMap   = make(map[string]string)
	mu       sync.RWMutex
	tokenLen = 10
)

// EncodeURL stores the URL and returns a short token (10 characters)
func EncodeURL(url string) string {
	token := generateShortToken(url)

	mu.Lock()
	urlMap[token] = url
	mu.Unlock()

	return token
}

// DecodeURL retrieves the original URL using the token
func DecodeURL(token string) (string, error) {
	mu.RLock()
	url, ok := urlMap[token]
	mu.RUnlock()

	if !ok {
		return "", errors.New("token not found")
	}
	return url, nil
}

// generateShortToken creates a consistent 10-character hash from the URL
func generateShortToken(url string) string {
	hash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(hash[:])[:tokenLen]
}
