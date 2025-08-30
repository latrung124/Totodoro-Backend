/*
File: internal/api_gateway/authentication/middleware.go
Author: trung.la
Date: 08/30/2025
Description: Middleware Authentication for using OIDCAuthenticator to authenticate requests
*/

package middleware

import (
	"context"
	"net/http"
	"strings"

	oidc "github.com/latrung124/Totodoro-Backend/internal/api_gateway/authentication/oidc"
)

type keyType string

const (
	contextUserID keyType = "user_id"
	contextEmail  keyType = "email"
)

// TokenVerifier is an interface for testability.
type TokenVerifier interface {
	VerifyIDToken(ctx context.Context, token string) (*oidc.GoogleClaims, error)
}

// AuthMiddleware is a middleware that authenticates requests using the provided TokenVerifier.
func AuthMiddleware(authenticator TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Allow health without auth
			if r.URL.Path == "/healthz" {
				next.ServeHTTP(w, r)
				return
			}

			authz := r.Header.Get("Authorization")
			if authz == "" {
				http.Error(w, "missing Authorization header", http.StatusUnauthorized)
				return
			}
			parts := strings.SplitN(authz, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, "invalid Authorization format", http.StatusUnauthorized)
				return
			}

			claims, err := authenticator.VerifyIDToken(r.Context(), parts[1])
			if err != nil {
				http.Error(w, "unauthorized: "+err.Error(), http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), contextUserID, claims.Sub)
			ctx = context.WithValue(ctx, contextEmail, claims.Email)
			r.Header.Set("X-User-Id", claims.Sub)
			r.Header.Set("X-User-Email", claims.Email)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext extracts the user ID from the context.
func UserIDFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(contextUserID)
	id, ok := v.(string)
	return id, ok
}

// EmailFromContext extracts the email from the context.
func EmailFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(contextEmail)
	email, ok := v.(string)
	return email, ok
}
