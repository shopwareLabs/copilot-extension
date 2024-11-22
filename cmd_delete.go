package main

import (
	"github.com/shopwarelabs/copilot-extension/config"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete document from vector db",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		collection, err := config.GetCollection()

		if err != nil {
			return err
		}

		return collection.Delete(cmd.Context(), nil, nil, args...)
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
