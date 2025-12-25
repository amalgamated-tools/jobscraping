// Package main implements a CLI tool for scraping job postings from various ATS platforms.
package main

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"os"

	"github.com/amalgamated-tools/jobscraping/pkg/helpers"
	"github.com/buger/jsonparser"
	_ "modernc.org/sqlite"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	ctx := context.Background()

	companies := []string{
		"abnormalsecurity",
		"accrue",
		"acuitymd",
		"ada18",
		"affirm",
		"agebold",
		"agilityrobotics",
		"airbnb",
		"alpaca",
		"alt",
		"andurilindustries",
		"anthropic",
		"apolloio",
		"array",
		"assemblyai",
		"aura798",
		"baselayer",
		"baton",
		"beautifulai",
		"billiontoone",
		"bishopfox",
		"bitwarden",
		"blackforestlabs",
		"blend",
		"block",
		"bluefishai",
		"brave",
		"brex",
		"butlr",
		"calendly",
		"calm",
		"capitalrx",
		"carefeed",
		"careportalinc",
	}
	locations := make(map[string]bool)

	for _, company := range companies {
		// The URL is like https://boards-api.greenhouse.io/v1/boards/{companyName}/jobs?content=true
		companyURL := fmt.Sprintf("https://boards-api.greenhouse.io/v1/boards/%s/jobs?content=true&pay_transparency=true", company)

		// Get the JSON from the company job board endpoint
		body, err := helpers.GetJSON(ctx, companyURL, nil)
		if err != nil {
			slog.ErrorContext(ctx, "Error getting JSON from Greenhouse job board endpoint", slog.String("url", companyURL), slog.Any("error", err))
			continue
		}

		_, _ = jsonparser.ArrayEach(body, func(value []byte, _ jsonparser.ValueType, _ int, _ error) {
			location, err := jsonparser.GetString(value, "location", "name")
			if err != nil {
				slog.ErrorContext(ctx, "Error getting job title", slog.Any("error", err))
				return
			}

			locations[location] = true
			slog.InfoContext(ctx, "Job location", slog.String("company", company), slog.String("location", location))
		}, "jobs")

		// err = os.WriteFile(fmt.Sprintf("%s_jobs.json", company), body, 0644)
		// if err != nil {
		// 	slog.ErrorContext(ctx, "Error writing JSON to file", slog.String("file", fmt.Sprintf("%s_jobs.json", company)), slog.Any("error", err))
		// 	continue
		// }

		// slog.InfoContext(ctx, "Wrote jobs to file", slog.String("file", fmt.Sprintf("%s_jobs.json", company)))
	}

	// jsonLocations, err := json.MarshalIndent(locations, " ", "  ")
	// if err != nil {
	// 	slog.ErrorContext(ctx, "Error marshaling locations to JSON", slog.Any("error", err))
	// 	return
	// }

	// err = os.WriteFile("locations.json", jsonLocations, 0644)
	// if err != nil {
	// 	slog.ErrorContext(ctx, "Error writing locations JSON to file", slog.String("file", "locations.json"), slog.Any("error", err))
	// 	return
	// }
	// slog.InfoContext(ctx, "Scraped company", slog.Int("job_count", len(jobs)))

	// job, err := ashby.ScrapeJob(ctx, "ashby", "6765ef2e-7905-4fbc-b941-783049e7835f")
	// if err != nil {
	// 	panic(err)
	// }

	// slog.InfoContext(ctx, "Scraped job", slog.String("title", job.Title), slog.String("location", job.Location))

	// ab, err := sql.Open("sqlite", "file:db/jobscraping.db?cache=shared&mode=rwc")
	// if err != nil {
	// 	panic(err)
	// }

	// newJob := models.Job{
	// 	Source:           "example",
	// 	SourceID:         "12345",
	// 	City:             helpers.Ptr("Chesapeake"),
	// 	Country:          helpers.Ptr("USA"),
	// 	DatePosted:       time.Now(),
	// 	Department:       models.CustomerSuccessSupport,
	// 	Description:      "This is an example job description.",
	// 	EmploymentType:   models.FullTime,
	// 	Equity:           models.EquityNotOffered,
	// 	IsRemote:         true,
	// 	LocationAddress:  helpers.Ptr("123 Main St, Chesapeake, VA"),
	// 	LocationType:     models.OnsiteLocation,
	// 	CompensationUnit: helpers.Ptr("YEARLY"),
	// 	Title:            "Software Engineer",
	// 	MinCompensation:  60000,
	// 	MaxCompensation:  120000,
	// 	Tags: map[string][]string{
	// 		"skills": {"Go", "SQL", "Docker"},
	// 	},
	// }

	// jsonNewJob, err := json.MarshalIndent(newJob, " ", "  ")
	// if err != nil {
	// 	panic(err)
	// }

	// queries := db.New(ab)
	// job, err := queries.CreateJob(ctx, db.CreateJobParams{
	// 	AbsoluteUrl: "https://example.com/job/1",
	// 	Data:        string(jsonNewJob),
	// })
	// if err != nil {
	// 	panic(err)
	// }
	// _, err = queries.GetJob(ctx, job.ID)
	// if err != nil {
	// 	panic(err)
	// }
	// _, err = queries.ListJobs(ctx)
	// if err != nil {
	// 	panic(err)
	// }
}
