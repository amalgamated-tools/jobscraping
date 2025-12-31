// Package main implements a CLI tool for scraping job postings from various ATS platforms.
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/amalgamated-tools/jobscraping/pkg/ats/rippling"
	_ "modernc.org/sqlite"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	for _, company := range []string{
		"boom-supersonic",
		"360-fire-flood",
		"aalo-atomics",
		"agentsync-careers",
		"community-phone-careers",
		"samacare"} {
		rippling.ScrapeCompany(context.Background(), company)
	}
}
