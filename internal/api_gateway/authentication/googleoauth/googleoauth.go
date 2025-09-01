/*
File: internal/api_gateway/authentication/googleoauth/googleoauth.go
Author: trung.la
Date: 09/01/2025
Description: Google OAuth authentication implementation
*/

package googleoauth

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	IDToken      string    `json:"id_token,omitempty"`
	TokenType    string    `json:"token_type"`
	Expiry       time.Time `json:"expiry"`
}

// NewConfig builds an OAuth2 config for Google web flow.
// scopes should include "openid", "email", "profile" if you need an ID token.
func NewConfig(clientID, clientSecret, redirectURL string, scopes []string) (*oauth2.Config, error) {
	if clientID == "" || clientSecret == "" || redirectURL == "" {
		return nil, fmt.Errorf("clientID, clientSecret, and redirectURL are required")
	}
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
		Endpoint:     google.Endpoint,
	}, nil
}

// ExchangeCode exchanges an auth code for tokens. Pass codeVerifier if using PKCE.
func ExchangeCode(ctx context.Context, cfg *oauth2.Config, code, codeVerifier string) (*TokenResponse, error) {
	var opts []oauth2.AuthCodeOption
	if codeVerifier != "" {
		opts = append(opts, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	}
	tok, err := cfg.Exchange(ctx, code, opts...)
	if err != nil {
		return nil, fmt.Errorf("exchange failed: %w", err)
	}
	resp := &TokenResponse{
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		TokenType:    tok.TokenType,
		Expiry:       tok.Expiry,
	}
	if idRaw, ok := tok.Extra("id_token").(string); ok {
		resp.IDToken = idRaw
	}
	return resp, nil
}

// Refresh uses a refresh_token to get a new access token.
func Refresh(ctx context.Context, cfg *oauth2.Config, refreshToken string) (*TokenResponse, error) {
	src := cfg.TokenSource(ctx, &oauth2.Token{RefreshToken: refreshToken})
	tok, err := src.Token()
	if err != nil {
		return nil, fmt.Errorf("refresh failed: %w", err)
	}
	resp := &TokenResponse{
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken, // may be empty on refresh
		TokenType:    tok.TokenType,
		Expiry:       tok.Expiry,
	}
	if idRaw, ok := tok.Extra("id_token").(string); ok {
		resp.IDToken = idRaw
	}
	return resp, nil
}
