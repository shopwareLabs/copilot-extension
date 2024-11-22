package config

import (
	"fmt"
	"os"
)

type Info struct {
	// Port is the local port on which the application will run
	Port string

	// FQDN (for Fully-Qualified Domain Name) is the internet facing host address
	// where application will live (e.g. https://example.com)
	FQDN string

	// ClientID comes from your configured GitHub app
	ClientID string

	// ClientSecret comes from your configured GitHub app
	ClientSecret string

	// LocalGitHubToken is a personal access token for the GitHub API. Used for index command
	LocalGitHubToken string

	// LocalGitHubIntegrationID is the integration ID for the GitHub app
	LocalGitHubIntegrationID string
}

const (
	clientIdEnv      = "CLIENT_ID"
	clientSecretEnv  = "CLIENT_SECRET"
	fqdnEnv          = "FQDN"
	localGitHubToken = "LOCAL_GITHUB_TOKEN"
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

	return &Info{
		Port:                     "8080",
		FQDN:                     fqdn,
		ClientID:                 clientID,
		ClientSecret:             clientSecret,
		LocalGitHubToken:         os.Getenv(localGitHubToken),
		LocalGitHubIntegrationID: "679",
	}, nil
}
