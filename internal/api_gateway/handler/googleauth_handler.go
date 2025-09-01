/*
File: internal/api_gateway/authentication/googleoauth/googleoauth.go
Author: trung.la
Date: 09/01/2025
*Description: Google OAuth Handler exchanging auth codes for tokens.
*/

package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/latrung124/Totodoro-Backend/internal/api_gateway/authentication/googleoauth"
	"golang.org/x/oauth2"
)

type GoogleOAuthDeps interface {
	NewConfig(clientID, clientSecret, redirectURL string, scopes []string) (*oauth2.Config, error)
	ExchangeCode(ctx context.Context, cfg *oauth2.Config, code, codeVerifier string) (*googleoauth.TokenResponse, error)
}

// Simple handler to exchange code -> tokens.
type GoogleAuthHandler struct {
	ClientID     string
	ClientSecret string
	Scopes       []string
}

type ExchangeReq struct {
	Code         string `json:"code"`
	RedirectURI  string `json:"redirectUri"`
	CodeVerifier string `json:"codeVerifier,omitempty"`
}

func (h *GoogleAuthHandler) RegisterGoogleAuthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/auth/google/exchange", h.exchange)
}

// Exchange authorization code for tokens.
func (h *GoogleAuthHandler) exchange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req ExchangeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" || req.RedirectURI == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	cfg, err := googleoauth.NewConfig(h.ClientID, h.ClientSecret, req.RedirectURI, []string{"openid", "email", "profile"})
	if err != nil {
		http.Error(w, "config error", http.StatusInternalServerError)
		return
	}
	tok, err := googleoauth.ExchangeCode(r.Context(), cfg, req.Code, req.CodeVerifier)
	if err != nil {
		http.Error(w, "exchange failed: "+err.Error(), http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tok)
}
