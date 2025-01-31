package config

import (
	"fmt"
	"os"
)

type Info struct {
	// FQDN (for Fully-Qualified Domain Name) is the internet facing host address
	// where application will live (e.g. https://example.com)
	FQDN string

	// ClientID comes from your configured GitHub app
	ClientID string

	// ClientSecret comes from your configured GitHub app
	ClientSecret string

	// OllamaHost is the host address of the Ollama API
	OllamaHost string
}

const (
	clientIdEnv     = "CLIENT_ID"
	clientSecretEnv = "CLIENT_SECRET"
	fqdnEnv         = "FQDN"
	ollamaHost      = "OLLAMA_HOST"
)

func New() (*Info, error) {
	fqdn := os.Getenv(fqdnEnv)
	if fqdn == "" {
		return nil, fmt.Errorf("%s environment variable required", fqdnEnv)
	}

	clientID := os.Getenv(clientIdEnv)
	if clientID == "" {
		return nil, fmt.Errorf("%s environment variable required", clientIdEnv)
	}

	clientSecret := os.Getenv(clientSecretEnv)
	if clientSecret == "" {
		return nil, fmt.Errorf("%s environment variable required", clientSecretEnv)
	}

	ollamaHost := os.Getenv(ollamaHost)
	if ollamaHost == "" {
		ollamaHost = "http://localhost:11434/api"
	}

	return &Info{
		FQDN:         fqdn,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		OllamaHost:   ollamaHost,
	}, nil
}
