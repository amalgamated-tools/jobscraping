// Package main implements a CLI tool for scraping job postings from various ATS platforms.
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/amalgamated-tools/jobscraping/pkg/ats/ashby"
	"github.com/amalgamated-tools/jobscraping/pkg/ats/bamboo"
	"github.com/amalgamated-tools/jobscraping/pkg/ats/gem"
	"github.com/amalgamated-tools/jobscraping/pkg/ats/greenhouse"
	_ "modernc.org/sqlite"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	// ashbyExample()
	// bambooExample()
	// gemExample()
	greenhouseExample()
}

func ashbyExample() {
	slog.Info("Starting Ashby ATS scraping example")

	jobs, err := ashby.ScrapeCompany(context.Background(), "acorns")
	if err != nil {
		slog.Error("failed to scrape jobs", slog.Any("error", err))
		os.Exit(1)
	}

	for _, job := range jobs {
		slog.Info("job found", slog.String("title", job.Title), slog.String("location", job.Location), slog.String("url", job.URL))
	}

	job, err := ashby.ScrapeJob(context.Background(), "acorns", jobs[0].SourceID)
	if err != nil {
		slog.Error("failed to scrape job", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("job found", slog.String("title", job.Title), slog.String("location", job.Location), slog.String("url", job.URL))
}

func bambooExample() {
	slog.Info("Starting Bamboo ATS scraping example")

	jobs, err := bamboo.ScrapeCompany(context.Background(), "axiscare")
	if err != nil {
		slog.Error("failed to scrape jobs", slog.Any("error", err))
		os.Exit(1)
	}

	for _, job := range jobs {
		slog.Info("job found", slog.String("title", job.Title), slog.String("location", job.Location), slog.String("url", job.URL))
	}

	job, err := bamboo.ScrapeJob(context.Background(), "axiscare", jobs[0].SourceID)
	if err != nil {
		slog.Error("failed to scrape job", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("job found", slog.String("title", job.Title), slog.String("location", job.Location), slog.String("url", job.URL))
}

func gemExample() {
	slog.Info("Starting Gem ATS scraping example")

	jobs, err := gem.ScrapeCompany(context.Background(), "bluesky")
	if err != nil {
		slog.Error("failed to scrape jobs", slog.Any("error", err))
		os.Exit(1)
	}

	for _, job := range jobs {
		slog.Info("job found", slog.String("title", job.Title), slog.String("location", job.Location), slog.String("url", job.URL))
	}

	job, err := gem.ScrapeJob(context.Background(), "bluesky", jobs[0].SourceID)
	if err != nil {
		slog.Error("failed to scrape job", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("job found", slog.String("title", job.Title), slog.String("location", job.Location), slog.String("url", job.URL))
}

func greenhouseExample() {
	slog.Info("Starting Greenhouse ATS scraping example")

	jobs, err := greenhouse.ScrapeCompany(context.Background(), "calendly")
	if err != nil {
		slog.Error("failed to scrape jobs", slog.Any("error", err))
		os.Exit(1)
	}

	for _, job := range jobs {
		slog.Info("job found", slog.String("title", job.Title), slog.String("location", job.Location), slog.String("url", job.URL))
	}

	job, err := greenhouse.ScrapeJob(context.Background(), "calendly", jobs[0].SourceID)
	if err != nil {
		slog.Error("failed to scrape job", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("job found", slog.String("title", job.Title), slog.String("location", job.Location), slog.String("url", job.URL))
}
