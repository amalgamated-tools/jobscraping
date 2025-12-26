// Package main implements a CLI tool for scraping job postings from various ATS platforms.
package main

import (
	"context"
	_ "embed"
	"log/slog"
	"os"

	"github.com/amalgamated-tools/jobscraping/pkg/ats/gem"
	_ "modernc.org/sqlite"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	gem.ScrapeCompany(context.Background(), "arlo")
	gem.ScrapeJob(context.Background(), "arlo", "am9icG9zdDruRXVwItfMiCqE7gmjrD4Q")
}
