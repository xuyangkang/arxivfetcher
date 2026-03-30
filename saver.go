package main

import "context"

type PaperSaver interface {
	Save(ctx context.Context, id, title, pdfPath string) (string, error)
}
