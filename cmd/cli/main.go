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
	"github.com/amalgamated-tools/jobscraping/pkg/ats/lever"
	"github.com/amalgamated-tools/jobscraping/pkg/ats/rippling"
	"github.com/amalgamated-tools/jobscraping/pkg/ats/workable"
	_ "modernc.org/sqlite"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	// ashbyExample()
	// bambooExample()
	// gemExample()
	greenhouseExample()
	// leverExample()
	// ripplingExample()
	// workableExample()
}

func ashbyExample() {
	slog.Info("Starting Ashby ATS scraping example")

	jobs, err := ashby.ScrapeCompany(context.Background(), "acorns")
	if err != nil {
		slog.Error("failed to scrape jobs", slog.Any("error", err))
		os.Exit(1)
	}

	for _, job := range jobs {
		slog.Info("job found", slog.String("title", job.Title), slog.String("location", job.Location), slog.String("url", job.URL), slog.String("company", job.Company.Name))
	}

	job, err := ashby.ScrapeJob(context.Background(), "acorns", jobs[0].SourceID)
	if err != nil {
		slog.Error("failed to scrape job", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("job found", slog.String("title", job.Title), slog.String("location", job.Location), slog.String("url", job.URL), slog.String("company", job.Company.Name))
}

func bambooExample() {
	slog.Info("Starting Bamboo ATS scraping example")

	jobs, err := bamboo.ScrapeCompany(context.Background(), "axiscare")
	if err != nil {
		slog.Error("failed to scrape jobs", slog.Any("error", err))
		os.Exit(1)
	}

	for _, job := range jobs {
		slog.Info("job found", slog.String("title", job.Title), slog.String("location", job.Location), slog.String("url", job.URL), slog.String("company", job.Company.Name))
	}

	job, err := bamboo.ScrapeJob(context.Background(), "axiscare", jobs[0].SourceID)
	if err != nil {
		slog.Error("failed to scrape job", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("job found", slog.String("title", job.Title), slog.String("location", job.Location), slog.String("url", job.URL), slog.String("company", job.Company.Name))
}

func gemExample() {
	slog.Info("Starting Gem ATS scraping example")

	jobs, err := gem.ScrapeCompany(context.Background(), "flint")
	if err != nil {
		slog.Error("failed to scrape jobs", slog.Any("error", err))
		os.Exit(1)
	}

	for _, job := range jobs {
		slog.Info("job found", slog.String("title", job.Title), slog.String("location", job.Location), slog.String("url", job.URL), slog.String("company", job.Company.Name))
	}

	job, err := gem.ScrapeJob(context.Background(), "flint", jobs[0].SourceID)
	if err != nil {
		slog.Error("failed to scrape job", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("job found", slog.String("title", job.Title), slog.String("location", job.Location), slog.String("url", job.URL), slog.String("company", job.Company.Name))
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

func leverExample() {
	slog.Info("Starting Lever ATS scraping example")

	jobs, err := lever.ScrapeCompany(context.Background(), "airalo")
	if err != nil {
		slog.Error("failed to scrape jobs", slog.Any("error", err))
		os.Exit(1)
	}

	for _, job := range jobs {
		slog.Info("job found", slog.String("title", job.Title), slog.String("location", job.Location), slog.String("url", job.URL))
	}

	job, err := lever.ScrapeJob(context.Background(), "airalo", jobs[0].SourceID)
	if err != nil {
		slog.Error("failed to scrape job", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("job found", slog.String("title", job.Title), slog.String("location", job.Location), slog.String("url", job.URL))
}

func ripplingExample() {
	slog.Info("Starting Rippling ATS scraping example")

	jobs, err := rippling.ScrapeCompany(context.Background(), "meroka")
	if err != nil {
		slog.Error("failed to scrape jobs", slog.Any("error", err))
		os.Exit(1)
	}

	for _, job := range jobs {
		slog.Info("job found", slog.String("title", job.Title), slog.String("location", job.Location), slog.String("url", job.URL))
	}

	job, err := rippling.ScrapeJob(context.Background(), "meroka", jobs[0].SourceID)
	if err != nil {
		slog.Error("failed to scrape job", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("job found", slog.String("title", job.Title), slog.String("location", job.Location), slog.String("url", job.URL))
}

func workableExample() {
	slog.Info("Starting Workable ATS scraping example")

	jobs, err := workable.ScrapeCompany(context.Background(), "avantstay")
	if err != nil {
		slog.Error("failed to scrape jobs", slog.Any("error", err))
		os.Exit(1)
	}

	for _, job := range jobs {
		slog.Info("job found", slog.String("title", job.Title), slog.String("location", job.Location), slog.String("url", job.URL))
	}

	job, err := workable.ScrapeJob(context.Background(), "avantstay", jobs[0].SourceID)
	if err != nil {
		slog.Error("failed to scrape job", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("job found", slog.String("title", job.Title), slog.String("location", job.Location), slog.String("url", job.URL))
}
