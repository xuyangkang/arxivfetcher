package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
)

// downloadPDF fetches a PDF from a given arxiv ID URL.
func downloadPDF(entryID string, destPath string) error {
	// The entryID usually looks like "http://arxiv.org/abs/2403.01234v1"
	// We convert it to "https://arxiv.org/pdf/2403.01234v1.pdf"
	
	idStr := path.Base(entryID) // "2403.01234v1"
	pdfURL := fmt.Sprintf("https://arxiv.org/pdf/%s.pdf", idStr)

	// Since we are creating a PDF downloader, let's execute the request
	resp, err := http.Get(pdfURL)
	if err != nil {
		return fmt.Errorf("failed to make request to %s: %v", pdfURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file at %s: %v", destPath, err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to copy pdf content: %v", err)
	}

	return nil
}
