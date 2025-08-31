/*
File: middleware_test.go
Author: trung.la
Date: 08/31/2025
Description: Unit tests for authentication middleware.
*/

package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	oidc "github.com/latrung124/Totodoro-Backend/internal/api_gateway/authentication/oidc"
)

type fakeVerifier struct {
	wantToken string
	claims    *oidc.GoogleClaims
	err       error
}

func (f fakeVerifier) VerifyIDToken(ctx context.Context, token string) (*oidc.GoogleClaims, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.wantToken != "" && token != f.wantToken {
		return nil, errors.New("unexpected token")
	}
	return f.claims, nil
}

func TestAuthMiddleware_HealthzBypass(t *testing.T) {
	verifier := fakeVerifier{err: errors.New("should not be called")}
	h := AuthMiddleware(verifier)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Should reach here without auth
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /healthz, got %d", rr.Code)
	}
}

func TestAuthMiddleware_MissingAuthHeader(t *testing.T) {
	verifier := fakeVerifier{}
	h := AuthMiddleware(verifier)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/123", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing Authorization, got %d", rr.Code)
	}
}

func TestAuthMiddleware_BadFormat(t *testing.T) {
	verifier := fakeVerifier{}
	h := AuthMiddleware(verifier)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/123", nil)
	req.Header.Set("Authorization", "Bearer") // no token part
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for bad Authorization format, got %d", rr.Code)
	}
}

func TestAuthMiddleware_VerifyError(t *testing.T) {
	verifier := fakeVerifier{wantToken: "bad", err: errors.New("invalid token")}
	h := AuthMiddleware(verifier)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/123", nil)
	req.Header.Set("Authorization", "Bearer bad")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 when verifier fails, got %d", rr.Code)
	}
}

func TestAuthMiddleware_Success(t *testing.T) {
	claims := &oidc.GoogleClaims{Sub: "u1", Email: "u1@example.com"}
	verifier := fakeVerifier{wantToken: "ok", claims: claims}

	// Handler asserts context and request headers are populated
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id, ok := UserIDFromContext(r.Context()); !ok || id != claims.Sub {
			t.Fatalf("user id missing or wrong in context: %v", id)
		}
		if email, ok := EmailFromContext(r.Context()); !ok || email != claims.Email {
			t.Fatalf("email missing or wrong in context: %v", email)
		}
		if got := r.Header.Get("X-User-Id"); got != claims.Sub {
			t.Fatalf("X-User-Id header not set, got %q", got)
		}
		if got := r.Header.Get("X-User-Email"); got != claims.Email {
			t.Fatalf("X-User-Email header not set, got %q", got)
		}
		w.WriteHeader(http.StatusOK)
	})

	h := AuthMiddleware(verifier)(next)

	// Case-insensitive scheme should work
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/123", nil)
	req.Header.Set("Authorization", "bearer ok")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 when auth succeeds, got %d", rr.Code)
	}
}
