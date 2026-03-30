package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type LocalSaver struct {
	BaseDir string
}

func (s *LocalSaver) Save(ctx context.Context, id, title, pdfPath string) (string, error) {
	if err := os.MkdirAll(s.BaseDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create base directory %s: %v", s.BaseDir, err)
	}

	destPath := filepath.Join(s.BaseDir, id+".pdf")

	src, err := os.Open(pdfPath)
	if err != nil {
		return "", fmt.Errorf("failed to open source pdf: %v", err)
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination pdf: %v", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to copy pdf content: %v", err)
	}

	return destPath, nil
}
