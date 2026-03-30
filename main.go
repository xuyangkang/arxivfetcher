package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/orijtech/arxiv/v1"
)

func fetchAndSummarize(keyword, outputDir string, maxResults int, gDriveFolder, credsFile, apiKey, baseURL, model string) {
	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Could not create output directory: %v", err)
	}

	historyFile := filepath.Join(outputDir, "history.txt")
	fetchedIDs := make(map[string]bool)

	file, err := os.Open(historyFile)
	if err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			id := strings.TrimSpace(scanner.Text())
			if id != "" {
				fetchedIDs[id] = true
			}
		}
		file.Close()
	}

	ctx, cancel := context.WithCancel(context.Background())
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
		log.Fatal(err)
	}

	var newIDs []string
	totalProcessed := 0
OuterLoop:
	for res := range resChan {
		if res.Err != nil {
			log.Printf("Error: %v", res.Err)
			continue
		}
		for _, entry := range res.Feed.Entry {
			if totalProcessed >= maxResults {
				cancel()
				break OuterLoop
			}
			totalProcessed++
			id := entry.ID
			if fetchedIDs[id] {
				continue
			}

			shortID := path.Base(id)
			pubTime, err := time.Parse(time.RFC3339, string(entry.Published))
			if err != nil {
				log.Printf("Could not parse publication date %s: %v", entry.Published, err)
				continue
			}

			content := fmt.Sprintf("Title: %s\n\nAbstract: %s\n", entry.Title, entry.Summary.Body)

			// AI Filter validation
			filterResp, filterErr := filterPaper(ctx, apiKey, baseURL, model, content)
			if filterErr != nil {
				log.Printf("Filter Error for %s: %v", id, filterErr)
				continue
			}

			if !filterResp.Match {
				fmt.Printf("Rejected: %s (Reason: %s)\n", shortID, filterResp.Justification)
				continue
			}

			fmt.Printf("Matched: %s (Reason: %s)\n", shortID, filterResp.Justification)

			// Store metadata / history
			dateDir := pubTime.Format("20060102")
			fullDateDir := filepath.Join(outputDir, dateDir)
			if err := os.MkdirAll(fullDateDir, 0755); err != nil {
				log.Printf("Could not create date directory %s: %v", fullDateDir, err)
				continue
			}

			paperFile := filepath.Join(fullDateDir, shortID+".txt")
			if err := os.WriteFile(paperFile, []byte(content), 0644); err != nil {
				log.Printf("Could not write paper file %s: %v", paperFile, err)
				continue
			}

			// Download PDF
			pdfFile := filepath.Join(fullDateDir, shortID+".pdf")
			fmt.Printf("Downloading PDF: %s\n", pdfFile)
			if err := downloadPDF(id, pdfFile); err != nil {
				log.Printf("PDF Download Error for %s: %v", id, err)
			} else {
				// Upload to Google Drive if credentials exist
				if _, statErr := os.Stat(credsFile); statErr == nil {
					fmt.Printf("Uploading %s to Google Drive...\n", shortID)
					link, uploadErr := uploadFileToDrive(ctx, credsFile, gDriveFolder, pdfFile, shortID+".pdf")
					if uploadErr != nil {
						log.Printf("Drive Upload Error for %s: %v", id, uploadErr)
					} else {
						fmt.Printf("Uploaded successfully: %s\n", link)
					}
				} else {
					log.Printf("Skipping Google Drive upload for %s: Credentials file '%s' not found.", id, credsFile)
				}
			}

			newIDs = append(newIDs, id)
		}
		time.Sleep(3 * time.Second)
	}

	if len(newIDs) > 0 {
		f, err := os.OpenFile(historyFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("Could not open history file for writing: %v", err)
			return
		}
		defer f.Close()
		for _, id := range newIDs {
			if _, err := f.WriteString(id + "\n"); err != nil {
				log.Printf("Could not write to history file: %v", err)
			}
		}
	}
}

func main() {
	home, _ := os.UserHomeDir()
	defaultOutputDir := filepath.Join(home, "arxiv")

	keyword := flag.String("keyword", "string algorithm", "Search keyword for ArXiv")
	outputDir := flag.String("output_dir", defaultOutputDir, "Directory to store history and papers")
	maxResults := flag.Int("max_results", 100, "Maximum number of results to fetch per page")
	gDriveFolder := flag.String("drive_folder", "paper", "Google Drive folder name to upload papers")
	credsFile := flag.String("credentials", filepath.Join(home, "credentials.json"), "Path to Google Drive credentials.json file")

	// Filter API flags
	apiKey := flag.String("grok_apikey", os.Getenv("GROK_API_KEY"), "API key for Grok (or set GROK_API_KEY env var)")
	baseURL := flag.String("grok_baseurl", "https://api.x.ai/v1", "Base URL for the API")
	model := flag.String("grok_model", "grok-4-1-fast-reasoning", "Model to use for completion")
	
	flag.Parse()

	// If no API key is specified (from flag or env), we might still want to warn, but let's let filter Paper error gracefully
	fetchAndSummarize(*keyword, *outputDir, *maxResults, *gDriveFolder, *credsFile, *apiKey, *baseURL, *model)
}
