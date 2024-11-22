package main

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/philippgille/chromem-go"
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

		config, err := config.New()
		if err != nil {
			return fmt.Errorf("error fetching config: %w", err)
		}

		me, err := url.Parse(config.FQDN)
		if err != nil {
			return fmt.Errorf("unable to parse HOST environment variable: %w", err)
		}

		db, err := chromem.NewPersistentDB("./db", true)

		if err != nil {
			return err
		}

		collection, err := db.GetOrCreateCollection("shopware_1", nil, nil)
		if err != nil {
			return err
		}

		me.Path = "auth/callback"

		oauthService := oauth.NewService(config.ClientID, config.ClientSecret, me.String())
		http.HandleFunc("/auth/authorization", oauthService.PreAuth)
		http.HandleFunc("/auth/callback", oauthService.PostAuth)

		agentService := agent.NewService(pubKey, collection)

		http.HandleFunc("/agent", agentService.ChatCompletion)

		fmt.Println("Listening on port", config.Port)
		return http.ListenAndServe(":"+config.Port, nil)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
