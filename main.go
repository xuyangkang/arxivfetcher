package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/orijtech/arxiv/v1"
	"golang.org/x/time/rate"
)

const (
	StatusNew       = "new"
	StatusUnrelated = "unrelated"
	StatusRelated   = "related"
	StatusUploaded  = "uploaded"
)

type HistoryEntry struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

func fetchAndSummarize(ctx context.Context, keyword string, maxResults int, backend StorageBackend, limiter *rate.Limiter, apiKey, baseURL, model string) {
	paperStatus := make(map[string]string)

	history, err := backend.LoadHistory(ctx)
	if err != nil {
		log.Printf("Warning: could not load history: %v", err)
	} else {
		for _, entry := range history {
			paperStatus[entry.ID] = entry.Status
		}
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create arxiv Query
	query := &arxiv.Query{
		Terms:             keyword,
		MaxResultsPerPage: 20, // Reasonable page size
		SortBy:            arxiv.SortBySubmittedDate,
		SortOrder:         arxiv.SortDescending,
	}

	// Make the API request
	resChan, _, err := arxiv.Search(ctx, query)
	if err != nil {
		log.Fatalf("ArXiv Search Initialization Error: %v", err)
	}

	totalProcessed := 0
OuterLoop:
	for res := range resChan {
		if res.Err != nil {
			log.Printf("ArXiv Pagination/Fetch Error: %v", res.Err)
			continue
		}
		for _, entry := range res.Feed.Entry {
			if totalProcessed >= maxResults {
				cancel()
				break OuterLoop
			}
			totalProcessed++
			id := entry.ID

			status := paperStatus[id]
			if status == StatusUploaded || status == StatusUnrelated {
				continue
			}

			shortID := path.Base(id)
			if status == "" || status == StatusNew {
				content := fmt.Sprintf("Title: %s\n\nAbstract: %s\n", entry.Title, entry.Summary.Body)

				// Rate limit before AI filter
				if err := limiter.Wait(ctx); err != nil {
					log.Printf("Rate limit error: %v", err)
					continue
				}

				// AI Filter validation
				filterResp, filterErr := filterPaper(ctx, apiKey, baseURL, model, content)
				if filterErr != nil {
					log.Printf("Filter Error for %s: %v", id, filterErr)
					continue
				}

				if !filterResp.Match {
					fmt.Printf("Rejected: %s (Reason: %s)\n", shortID, filterResp.Justification)
					backend.UpdateHistory(ctx, HistoryEntry{ID: id, Status: StatusUnrelated, Reason: filterResp.Justification})
					paperStatus[id] = StatusUnrelated
					continue
				}

				fmt.Printf("Matched: %s (Reason: %s)\n", shortID, filterResp.Justification)
				backend.UpdateHistory(ctx, HistoryEntry{ID: id, Status: StatusRelated, Reason: filterResp.Justification})
				paperStatus[id] = StatusRelated
				status = StatusRelated
			}

			if status == StatusRelated {
				pdfFile := filepath.Join(os.TempDir(), shortID+".pdf")
				
				// Rate limit before PDF download
				if err := limiter.Wait(ctx); err != nil {
					log.Printf("Rate limit error: %v", err)
					continue
				}

				// Download PDF to temp directory
				fmt.Printf("Downloading PDF to temporary location: %s\n", pdfFile)
				if err := downloadPDF(id, pdfFile); err != nil {
					log.Printf("PDF Download Error for %s: %v", id, err)
					continue
				}

				fmt.Printf("Saving %s...\n", shortID)
				link, uploadErr := backend.SavePaper(ctx, shortID, entry.Title, pdfFile)
				if uploadErr != nil {
					log.Printf("Save Error for %s: %v", id, uploadErr)
				} else {
					fmt.Printf("Saved successfully: %s\n", link)
					backend.UpdateHistory(ctx, HistoryEntry{ID: id, Status: StatusUploaded})
					paperStatus[id] = StatusUploaded
					
					// Clean up the temp file after successful upload
					os.Remove(pdfFile)
				}
			}
		}
		time.Sleep(3 * time.Second)
	}
}

func main() {
	home, _ := os.UserHomeDir()
	defaultOutputDir := filepath.Join(home, "arxiv")

	keyword := flag.String("keyword", "string algorithm", "Search keyword for ArXiv")
	outputDir := flag.String("output_dir", defaultOutputDir, "Directory to store history and papers")
	maxResults := flag.Int("max_results", 100, "Maximum number of results to fetch per page")
	papersDir := flag.String("papers_dir", filepath.Join(home, "papers"), "Local directory to store useful papers")
	rps := flag.Float64("rps", 0.5, "Requests per second allowed for AI and PDF downloads (default 0.5)")

	// Filter API flags
	apiKey := flag.String("grok_apikey", os.Getenv("GROK_API_KEY"), "API key for Grok (or set GROK_API_KEY env var)")
	baseURL := flag.String("grok_baseurl", "https://api.x.ai/v1", "Base URL for the API")
	model := flag.String("grok_model", "grok-4-1-fast-reasoning", "Model to use for completion")

	flag.Parse()

	limiter := rate.NewLimiter(rate.Limit(*rps), 1)
	backend := &LocalSaver{
		BaseDir:    *papersDir,
		HistoryDir: *outputDir,
	}

	// If no API key is specified (from flag or env), we might still want to warn, but let's let filter Paper error gracefully
	fetchAndSummarize(context.Background(), *keyword, *maxResults, backend, limiter, *apiKey, *baseURL, *model)
}
