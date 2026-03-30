package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type LocalSaver struct {
	BaseDir    string
	HistoryDir string
}

func (s *LocalSaver) SavePaper(ctx context.Context, id, title, localPdfPath string) (string, error) {
	if err := os.MkdirAll(s.BaseDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create base directory %s: %v", s.BaseDir, err)
	}

	destPath := filepath.Join(s.BaseDir, id+".pdf")

	src, err := os.Open(localPdfPath)
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

func (s *LocalSaver) LoadHistory(ctx context.Context) ([]HistoryEntry, error) {
	historyFile := filepath.Join(s.HistoryDir, "historyv2.txt")
	var entries []HistoryEntry

	file, err := os.Open(historyFile)
	if err != nil {
		if os.IsNotExist(err) {
			return entries, nil
		}
		return nil, fmt.Errorf("failed to open history file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var entry HistoryEntry
		if err := json.Unmarshal([]byte(line), &entry); err == nil {
			entries = append(entries, entry)
		}
	}

	return entries, nil
}

func (s *LocalSaver) UpdateHistory(ctx context.Context, entry HistoryEntry) error {
	if err := os.MkdirAll(s.HistoryDir, 0755); err != nil {
		return fmt.Errorf("failed to create history directory: %v", err)
	}

	historyFile := filepath.Join(s.HistoryDir, "historyv2.txt")
	f, err := os.OpenFile(historyFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open history file for writing: %v", err)
	}
	defer f.Close()

	b, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal history entry: %v", err)
	}

	if _, err := f.Write(append(b, '\n')); err != nil {
		return fmt.Errorf("failed to write to history file: %v", err)
	}

	return nil
}
