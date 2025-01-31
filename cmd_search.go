package main

import (
	"fmt"

	"github.com/shopwarelabs/copilot-extension/config"
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

		collection, err := config.GetCollection(cfg)

		if err != nil {
			return err
		}

		result, err := collection.Query(cmd.Context(), args[0], 20, nil, nil)

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
