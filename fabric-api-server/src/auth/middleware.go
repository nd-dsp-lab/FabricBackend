package auth

import (
	"context"
	"net/http"
	"strings"
)

// contextKey is the key type for storing auth info in context
type contextKey struct{}

// AuthContext holds authentication information extracted from the token
type AuthContext struct {
	Token     string
	SecretKey string
}

// BearerAuthMiddleware creates a standard HTTP middleware for Bearer token authentication
// This works with Chi router and is compatible with Huma v2
func BearerAuthMiddleware(tokenStore *TokenStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for docs and OpenAPI endpoints
			if r.URL.Path == "/docs" || r.URL.Path == "/openapi.json" || r.URL.Path == "/openapi.yaml" {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "Missing Authorization header"}`))
				return
			}

			// Extract Bearer token
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "Invalid Authorization header format. Expected: Bearer <token>"}`))
				return
			}

			token := parts[1]
			if token == "" {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "Token is empty"}`))
				return
			}

			// Validate token
			secretKey, exists := tokenStore.ValidateToken(token)
			if !exists {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "Invalid or unknown token"}`))
				return
			}

			// Create auth context and inject into request context
			authCtx := &AuthContext{
				Token:     token,
				SecretKey: secretKey,
			}
			newContext := context.WithValue(r.Context(), contextKey{}, authCtx)

			next.ServeHTTP(w, r.WithContext(newContext))
		})
	}
}

// GetAuthContext retrieves the AuthContext from the context
// Returns nil if no auth context is found
func GetAuthContext(ctx context.Context) *AuthContext {
	if authCtx, ok := ctx.Value(contextKey{}).(*AuthContext); ok {
		return authCtx
	}
	return nil
}
