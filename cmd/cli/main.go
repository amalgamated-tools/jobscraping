// Package main implements a CLI tool for scraping job postings from various ATS platforms.
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/amalgamated-tools/jobscraping/pkg/ats/ashby"
	"github.com/amalgamated-tools/jobscraping/pkg/ats/greenhouse"
	_ "modernc.org/sqlite"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	job, err := greenhouse.ScrapeJob(context.Background(), "harpergroup", "4997480008")
	if err != nil {
		slog.Error("failed to scrape job", "error", err)
		os.Exit(1)
	}
	slog.Info("job found", "title", job.Title, "location", job.Location, "url", job.URL)
	jobs, err := ashby.ScrapeCompany(context.Background(), "1password")
	if err != nil {
		slog.Error("failed to scrape jobs", "error", err)
		os.Exit(1)
	}
	for _, job := range jobs {
		slog.Info("job found", "title", job.Title, "location", job.Location, "url", job.URL)
	}
}
