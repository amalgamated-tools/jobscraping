// Package main implements a CLI tool for scraping job postings from various ATS platforms.
package main

import (
	_ "embed"
	"log/slog"
	"os"

	_ "modernc.org/sqlite"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

}
