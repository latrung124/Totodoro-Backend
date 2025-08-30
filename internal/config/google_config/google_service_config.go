/*
File: internal/config/google_config.go
Author: trung.la
Date: 08/30/2025
Description: Google Service Key management.
*/

package google_config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/latrung124/Totodoro-Backend/internal/config"
)

// Exported fields + JSON tags so decoding works
type GoogleServiceConfig struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
	UniverseDomain          string `json:"universe_domain"`
}

func GetGoogleServiceConfig() (*GoogleServiceConfig, error) {
	config.Load()

	cfg, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	jsonFile, err := os.Open(cfg.GoogleCredentialPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open google credential file: %w", err)
	}
	defer jsonFile.Close()

	var googleServiceConfig GoogleServiceConfig
	decoder := json.NewDecoder(jsonFile)
	if err := decoder.Decode(&googleServiceConfig); err != nil {
		return nil, fmt.Errorf("failed to decode google credential file: %w", err)
	}

	return &googleServiceConfig, nil
}
