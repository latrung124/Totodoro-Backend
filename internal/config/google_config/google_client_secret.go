/*
File: google_client_secret.go
Author: trung.la
Date: 08/30/2025
Description: Google OAuth2 client secret management.
*/

package google_config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/latrung124/Totodoro-Backend/internal/config"
)

type GoogleClientSecret struct {
	ClientID                string   `json:"client_id"`
	ProjectID               string   `json:"project_id"`
	AuthURI                 string   `json:"auth_uri"`
	TokenURI                string   `json:"token_uri"`
	AuthProviderX509CertURL string   `json:"auth_provider_x509_cert_url"`
	ClientSecret            string   `json:"client_secret"`
	RedirectURIs            []string `json:"redirect_uris"`
}

func GetGoogleClientSecret() (*GoogleClientSecret, error) {
	config.Load()
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	_ = cfg // reserved for future use (e.g., path in config)

	// Read path from env to avoid hard-coding; supports two common env names.
	path := os.Getenv("GOOGLE_CLIENT_SECRET_PATH")
	if path == "" {
		path = os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET_PATH")
	}
	if path == "" {
		return nil, fmt.Errorf("GOOGLE_CLIENT_SECRET_PATH is not set")
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open client secret file: %w", err)
	}
	defer f.Close()

	// Google client secret JSON is wrapped under "installed" (or sometimes "web").
	var envelope struct {
		Installed GoogleClientSecret `json:"installed"`
		Web       GoogleClientSecret `json:"web"`
	}
	if err := json.NewDecoder(f).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("decode client secret json: %w", err)
	}

	// Prefer "installed" if present, else "web".
	if envelope.Installed.ClientID != "" {
		return &envelope.Installed, nil
	}
	if envelope.Web.ClientID != "" {
		return &envelope.Web, nil
	}
	return nil, fmt.Errorf("client secret JSON missing 'installed' or 'web' credentials")
}
