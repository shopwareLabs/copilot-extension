package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/shopwarelabs/copilot-extension/agent"
	"github.com/shopwarelabs/copilot-extension/config"
	"github.com/shopwarelabs/copilot-extension/oauth"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts the server",
	RunE: func(cmd *cobra.Command, args []string) error {
		pubKey, err := fetchPublicKey()
		if err != nil {
			return fmt.Errorf("failed to fetch public key: %w", err)
		}

		cfg, err := config.New()
		if err != nil {
			return fmt.Errorf("error fetching config: %w", err)
		}

		me, err := url.Parse(cfg.FQDN)
		if err != nil {
			return fmt.Errorf("unable to parse HOST environment variable: %w", err)
		}

		collection, err := config.GetCollection(cfg)

		if err != nil {
			return fmt.Errorf("failed to get collection: %w", err)
		}

		me.Path = "auth/callback"

		oauthService := oauth.NewService(cfg.ClientID, cfg.ClientSecret, me.String())
		http.HandleFunc("/auth/authorization", oauthService.PreAuth)
		http.HandleFunc("/auth/callback", oauthService.PostAuth)

		agentService := agent.NewService(pubKey, collection, os.Getenv("DEBUG") == "true")

		http.HandleFunc("/agent", agentService.ChatCompletion)
		http.HandleFunc("/search", agent.NewSearchService(collection).Search)

		fmt.Println("Listening on port 8000")
		return http.ListenAndServe(":8000", nil)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
