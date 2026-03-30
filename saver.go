package main

import "context"

type StorageBackend interface {
	SavePaper(ctx context.Context, id, title, localPdfPath string) (string, error)
	LoadHistory(ctx context.Context) ([]HistoryEntry, error)
	UpdateHistory(ctx context.Context, entry HistoryEntry) error
}
