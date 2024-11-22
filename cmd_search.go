package main

import (
	"fmt"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/shopwarelabs/copilot-extension/config"
	"github.com/shopwarelabs/copilot-extension/copilot"
	"github.com/spf13/cobra"
)

var cmdSearch = &cobra.Command{
	Use:   "search",
	Short: "Search for documents",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.New()

		if err != nil {
			return err
		}

		collection, err := config.GetCollection()

		if err != nil {
			return err
		}

		embeddings, err := copilot.Embeddings(cmd.Context(), retryablehttp.NewClient(), cfg.LocalGitHubIntegrationID, cfg.LocalGitHubToken, &copilot.EmbeddingsRequest{
			Model: copilot.ModelEmbeddings,
			Input: []string{args[0]},
		})

		if err != nil {
			return err
		}

		result, err := collection.QueryEmbedding(cmd.Context(), embeddings.Data[0].Embedding, 20, nil, nil)

		if err != nil {
			return err
		}

		for _, r := range result {
			fmt.Printf("%f - %s\n", r.Similarity, r.ID)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(cmdSearch)
}
