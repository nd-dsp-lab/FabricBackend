package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// TokenStore manages token-to-secret mappings stored in a local JSON file
type TokenStore struct {
	mu       sync.RWMutex
	filePath string
	tokens   map[string]string // token -> secretKey
}

// tokenData represents the JSON structure for token storage
type tokenData struct {
	Tokens map[string]string `json:"tokens"`
}

// NewTokenStore creates a new TokenStore with the specified file path
func NewTokenStore(filePath string) *TokenStore {
	return &TokenStore{
		filePath: filePath,
		tokens:   make(map[string]string),
	}
}

// LoadTokens loads tokens from the JSON file
func (ts *TokenStore) LoadTokens() error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	data, err := os.ReadFile(ts.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, start with empty map
			ts.tokens = make(map[string]string)
			return nil
		}
		return fmt.Errorf("failed to read token file: %w", err)
	}

	var td tokenData
	if err := json.Unmarshal(data, &td); err != nil {
		return fmt.Errorf("failed to parse token file: %w", err)
	}

	ts.tokens = td.Tokens
	return nil
}

// SaveTokens saves the current tokens to the JSON file
func (ts *TokenStore) SaveTokens() error {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	td := tokenData{Tokens: ts.tokens}
	data, err := json.MarshalIndent(td, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tokens: %w", err)
	}

	if err := os.WriteFile(ts.filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}

// ValidateToken checks if a token exists and returns its secret key
func (ts *TokenStore) ValidateToken(token string) (secretKey string, exists bool) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	secretKey, exists = ts.tokens[token]
	return secretKey, exists
}

// AddToken adds a new token-secret pair to the store
func (ts *TokenStore) AddToken(token, secretKey string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.tokens == nil {
		ts.tokens = make(map[string]string)
	}
	ts.tokens[token] = secretKey
}

// GenerateToken generates a random token of the specified length
func GenerateToken(length int) string {
	return generateRandomString(length)
}

// GenerateSecret generates a random secret key of the specified length
func GenerateSecret(length int) string {
	return generateRandomString(length)
}

// generateRandomString generates a random base64 URL-encoded string
func generateRandomString(length int) string {
	// Calculate bytes needed for the desired output length
	// base64.RawURLEncoding produces 4/3 bytes per input byte
	bytesNeeded := (length * 3) / 4
	if (length*3)%4 != 0 {
		bytesNeeded++
	}

	b := make([]byte, bytesNeeded)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	encoded := base64.RawURLEncoding.EncodeToString(b)
	if len(encoded) > length {
		return encoded[:length]
	}
	return encoded
}

// GetTokenCount returns the number of tokens in the store
func (ts *TokenStore) GetTokenCount() int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return len(ts.tokens)
}
