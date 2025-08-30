/*
File: internal/api_gateway/authentication/oidc.go
Author: trung.la
Date: 08/30/2025
Description: OIDC authentication handler
*/

package oidc

import (
	"context"
	"errors"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
)

type OIDCAuthenticator struct {
	verifier *oidc.IDTokenVerifier
}

type GoogleClaims struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
}

// NewOIDCAuthenticator creates a new OIDCAuthenticator with the given client ID.
func NewOIDCAuthenticator(ctx context.Context, clientID string) (*OIDCAuthenticator, error) {
	if clientID == "" {
		return nil, errors.New("clientID is required")
	}
	provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}
	cfg := &oidc.Config{ClientID: clientID}
	return &OIDCAuthenticator{verifier: provider.Verifier(cfg)}, nil
}

// VerifyIDToken verifies the ID token and returns the claims if valid.
func (v *OIDCAuthenticator) VerifyIDToken(ctx context.Context, token string) (*GoogleClaims, error) {
	if token == "" {
		return nil, errors.New("token is required")
	}
	idToken, err := v.verifier.Verify(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token: %w", err)
	}
	var claims GoogleClaims
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %w", err)
	}
	if claims.Sub == "" || claims.Email == "" {
		return nil, errors.New("missing required claims")
	}
	return &claims, nil
}
