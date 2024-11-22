package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/philippgille/chromem-go"
	"github.com/shopwarelabs/copilot-extension/config"
	"github.com/shopwarelabs/copilot-extension/copilot"
	"github.com/spf13/cobra"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/textsplitter"
)

var indexCommand = &cobra.Command{
	Use:   "index",
	Short: "index",
	RunE: func(cmd *cobra.Command, args []string) error {
		collection, err := config.GetCollection()

		if err != nil {
			return err
		}

		cfg, err := config.New()

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

		client := retryablehttp.NewClient()

		client.Backoff = retryablehttp.DefaultBackoff
		client.CheckRetry = retryablehttp.DefaultRetryPolicy

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

			documents := make([]chromem.Document, 0)

			for idx, doc := range docs {
				var embeddings *copilot.EmbeddingsResponse

				for {
					embeddings, err = copilot.Embeddings(cmd.Context(), client, cfg.LocalGitHubIntegrationID, cfg.LocalGitHubToken, &copilot.EmbeddingsRequest{
						Model: copilot.ModelEmbeddings,
						Input: []string{doc.PageContent},
					})

					if err == nil {
						break
					}

					if err != nil {
						if strings.Contains(err.Error(), "429") {
							log.Infof("Rate limited slowing down")
							time.Sleep(time.Second * 3)
							continue
						}

						log.Infof("File: %s, error: %s", fileName, err.Error())
						err = nil
						break
					}
				}

				if embeddings != nil {
					for _, embedding := range embeddings.Data {
						documents = append(documents, chromem.Document{
							ID:        fmt.Sprintf("%s_%d", fileName, idx),
							Content:   doc.PageContent,
							Embedding: embedding.Embedding,
						})
					}
				} else {
					break
				}
			}

			if len(documents) == 0 {
				continue
			}

			if err := collection.AddDocuments(cmd.Context(), documents, runtime.NumCPU()); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(indexCommand)
}
