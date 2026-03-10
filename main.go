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
    // Add PDF extractor, LLM client
)

func fetchAndSummarize(keyword string, outputDir string, maxResults int) {
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
    query := &arxiv.Query{
        Terms:             keyword,
        MaxResultsPerPage: 20, // Reasonable page size
        SortBy:            arxiv.SortBySubmittedDate,
        SortOrder:         arxiv.SortDescending,
    }
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

            // Prepare date-specific directory based on publication date
            pubTime, err := time.Parse(time.RFC3339, string(entry.Published))
            if err != nil {
                log.Printf("Could not parse publication date %s: %v", entry.Published, err)
                continue
            }
            dateDir := pubTime.Format("20060102")
            fullDateDir := filepath.Join(outputDir, dateDir)

            // Create date directory on demand
            if err := os.MkdirAll(fullDateDir, 0755); err != nil {
                log.Printf("Could not create date directory %s: %v", fullDateDir, err)
                continue
            }

            // Save paper details to file
            shortID := path.Base(id)
            paperFile := filepath.Join(fullDateDir, shortID+".txt")
            content := fmt.Sprintf("Title: %s\n\nAbstract: %s\n", entry.Title, entry.Summary.Body)
            if err := os.WriteFile(paperFile, []byte(content), 0644); err != nil {
                log.Printf("Could not write paper file %s: %v", paperFile, err)
                continue
            }

            fmt.Printf("Fetched: %s\n", shortID)
            newIDs = append(newIDs, id)
        }
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
    flag.Parse()

    fetchAndSummarize(*keyword, *outputDir, *maxResults)
}
