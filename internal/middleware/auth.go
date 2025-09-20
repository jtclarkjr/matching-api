package middleware

import (
	"context"
	"net/http"
	"os"

	"matching-api/internal/models"
	"matching-api/pkg/auth"
	"matching-api/pkg/utils"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	UserContextKey contextKey = "user"
)

// AuthMiddleware validates JWT tokens and sets user context
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get JWT secret from environment
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			jwtSecret = "your-super-secret-key-change-this-in-production"
		}

		jwtService := auth.NewJWTService(jwtSecret)

		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.WriteErrorResponse(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		token := auth.ExtractTokenFromHeader(authHeader)
		if token == "" {
			utils.WriteErrorResponse(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		// Validate token
		claims, err := jwtService.ValidateAccessToken(token)
		if err != nil {
			utils.WriteErrorResponse(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserFromContext extracts user claims from request context
func GetUserFromContext(ctx context.Context) (*models.JWTClaims, bool) {
	user, ok := ctx.Value(UserContextKey).(*models.JWTClaims)
	return user, ok
}