// Package main implements a CLI tool for scraping job postings from various ATS platforms.
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/amalgamated-tools/jobscraping/pkg/ats/bamboo"
	_ "modernc.org/sqlite"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	bamboo.ScrapeCompany(context.Background(), "beehiiv")
}
