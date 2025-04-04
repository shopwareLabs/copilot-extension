package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/philippgille/chromem-go"
	"github.com/shopwarelabs/copilot-extension/config"
	"github.com/spf13/cobra"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/textsplitter"
)

var (
	workers int
)

type indexJob struct {
	fileName string
	index    int
	total    int
}

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
		err = filepath.WalkDir("data", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
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

		if err != nil {
			return fmt.Errorf("failed to walk directory: %w", err)
		}

		if len(fileNames) == 0 {
			return fmt.Errorf("no files found to index")
		}

		split := textsplitter.NewRecursiveCharacter()
		split.ChunkSize = 12000
		split.ChunkOverlap = 30

		filesCount := len(fileNames)
		jobs := make(chan indexJob)
		errChan := make(chan error)
		var wg sync.WaitGroup

		re := regexp.MustCompile(`(?s)^---\n.*?---\n`)

		// Start worker pool
		for w := 1; w <= workers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for job := range jobs {
					content, err := os.Open(job.fileName)
					if err != nil {
						errChan <- fmt.Errorf("failed to open file %s: %w", job.fileName, err)
						continue
					}

					p := documentloaders.NewText(content)
					docs, err := p.LoadAndSplit(cmd.Context(), split)
					content.Close()

					if err != nil {
						errChan <- fmt.Errorf("failed to split file %s: %w", job.fileName, err)
						continue
					}

					log.Infof("Indexing [%d/%d] %s", job.index+1, job.total, job.fileName)

					for idx, doc := range docs {
						documentId := fmt.Sprintf("%s_%d", job.fileName, idx)

						lookupDoc, err := collection.GetByID(cmd.Context(), documentId)

						pathSplit := strings.Split(job.fileName, "/")
						source := pathSplit[1]

						if err != nil || lookupDoc.Content != doc.PageContent {
							if err := collection.AddDocument(cmd.Context(), chromem.Document{
								ID:      fmt.Sprintf("%s_%d", job.fileName, idx),
								Content: re.ReplaceAllString(doc.PageContent, ""),
								Metadata: map[string]string{
									"source": source,
									"file":   job.fileName,
								},
							}); err != nil {
								log.Error("failed to index document", "error", err)
								continue
							}
						}
					}
				}
			}()
		}

		// Error collector
		go func() {
			wg.Wait()
			close(errChan)
		}()

		// Send jobs to workers
		go func() {
			for i, fileName := range fileNames {
				jobs <- indexJob{
					fileName: fileName,
					index:    i,
					total:    filesCount,
				}
			}
			close(jobs)
		}()

		// Wait for errors
		for err := range errChan {
			log.Error("worker error", "error", err)
		}

		return nil
	},
}

func init() {
	indexCommand.Flags().IntVarP(&workers, "workers", "w", 4, "Number of parallel workers")
	rootCmd.AddCommand(indexCommand)
}
