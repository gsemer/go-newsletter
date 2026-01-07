package http

import (
	"context"
	"log/slog"
	"net/http"
	"newsletter/config"
	"newsletter/internal/users/domain"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Validate is a middleware that verifies the JWT access token for incoming requests.
//
// It checks the "Authorization" header for a Bearer token, validates the token,
// and extracts the user ID from its claims. If the token is valid, the middleware
// stores the user ID in the request context under `domain.UserID` and calls the next handler.
//
// On failure, it returns an HTTP 401 Unauthorized response for invalid tokens or
// missing bearer tokens, and HTTP 500 Internal Server Error if the JWT secret is not configured.
//
// Usage:
//
//	http.Handle("/protected", app.Validate(protectedHandler))
func (app *App) Validate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bearer := r.Header.Get("Authorization")

		if !strings.HasPrefix(bearer, "Bearer ") {
			http.Error(w, "no bearer token", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimSpace(strings.TrimPrefix(bearer, "Bearer "))

		secret := config.GetEnv("JWT_SECRET_KEY", "")
		if secret == "" {
			slog.Error("JWT secret is not set")
			http.Error(w, "server configuration error", http.StatusInternalServerError)
			return
		}

		token, err := jwt.ParseWithClaims(
			tokenString,
			&domain.Claims{},
			func(t *jwt.Token) (any, error) {
				return []byte(secret), nil
			},
		)
		if err != nil || !token.Valid {
			slog.Warn("invalid token", "error", err)
			http.Error(w, "token invalid", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*domain.Claims)
		if !ok || claims == nil {
			http.Error(w, "invalid claims", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), domain.UserID, claims.Subject)

		slog.Debug("authorized request", "user_id", claims.Subject, "path", r.URL.Path)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
