// Package main implements a CLI tool for scraping job postings from various ATS platforms.
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/amalgamated-tools/jobscraping/pkg/ats/ashby"
	_ "modernc.org/sqlite"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	jobs, err := ashby.ScrapeCompany(context.Background(), "1password")
	if err != nil {
		slog.Error("Error scraping jobs from Ashby", slog.Any("error", err))
		return
	}

	for _, job := range jobs {
		slog.Info("Scraped job", slog.String("title", job.Title), slog.String("company", job.Company.Name))
	}
}
