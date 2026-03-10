package main

import (
    "context"
    "flag"
    "fmt"
    "log"
    "github.com/orijtech/arxiv/v1"
    "github.com/robfig/cron/v3"
    // Add PDF extractor, LLM client
)

func fetchAndSummarize() {
    ctx := context.Background()
    query := &arxiv.Query{
        Terms:             "string algorithm",
        MaxResultsPerPage: 100,
        SortBy:            arxiv.SortBySubmittedDate,
        SortOrder:         arxiv.SortDescending,
    }
    resChan, _, err := arxiv.Search(ctx, query)
    if err != nil {
        log.Fatal(err)
    }
    for res := range resChan {
        if res.Err != nil {
            log.Printf("Error: %v", res.Err)
            continue
        }
        for _, entry := range res.Feed.Entry {
            fmt.Printf("Title: %s\nAbstract: %s\n", entry.Title, entry.Summary.Body)
            // Download PDF: entry.Link, extract text, call LLM for summary
        }
    }
}

func main() {
    runOnce := flag.Bool("run_once", false, "Run the fetcher once and exit")
    flag.Parse()

    if *runOnce {
        fetchAndSummarize()
        return
    }

    c := cron.New()
    c.AddFunc("0 8 * * *", fetchAndSummarize)  // Daily 8AM
    c.Start()
    select {}  // Run forever
}
