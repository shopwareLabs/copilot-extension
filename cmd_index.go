package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/philippgille/chromem-go"
	"github.com/shopwarelabs/copilot-extension/config"
	"github.com/spf13/cobra"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/textsplitter"
)

var indexCommand = &cobra.Command{
	Use:   "index",
	Short: "Embed all files in the data directory to the vector database",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.New()

		if err != nil {
			return err
		}

		collection, err := config.GetCollection(cfg)

		if err != nil {
			return err
		}

		fileNames := make([]string, 0)

		filepath.WalkDir("data", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}

			if d.IsDir() {
				return nil
			}

			if strings.Contains(path, "draco") {
				return nil
			}

			if strings.HasSuffix(d.Name(), ".md") || strings.HasSuffix(d.Name(), ".js") || strings.HasSuffix(d.Name(), ".php") || strings.HasSuffix(d.Name(), ".scss") || strings.HasSuffix(d.Name(), ".css") || strings.HasSuffix(d.Name(), ".twig") {
				fileNames = append(fileNames, path)
			}

			return nil
		})

		split := textsplitter.NewRecursiveCharacter()
		split.ChunkSize = 12000
		split.ChunkOverlap = 30

		filesCount := len(fileNames)

		for i, fileName := range fileNames {
			content, err := os.Open(fileName)

			if err != nil {
				continue
			}

			defer content.Close()

			p := documentloaders.NewText(content)

			docs, err := p.LoadAndSplit(cmd.Context(), split)

			if err != nil {
				continue
			}

			log.Infof("Indexing [%d/%d] %s", i+1, filesCount, fileName)

			for idx, doc := range docs {
				if err := collection.AddDocument(cmd.Context(), chromem.Document{
					ID:      fmt.Sprintf("%s_%d", fileName, idx),
					Content: doc.PageContent,
				}); err != nil {
					return err
				}

				log.Infof("Indexed [%d/%d] %s", idx+1, len(docs), fileName)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(indexCommand)
}
