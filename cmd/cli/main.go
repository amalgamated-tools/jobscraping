package main

import (
	"context"
	_ "embed"

	"github.com/amalgamated-tools/jobscraping/pkg/ats/ashby"
	_ "modernc.org/sqlite"
)

func main() {
	ctx := context.Background()
	// as an example, let's scrape the company "ashby" on Ashby
	// err := ashby.ScrapeCompany(ctx, "ashby")
	// if err != nil {
	// 	panic(err)
	// }
	err := ashby.ScrapeJob(ctx, "ashby", "6765ef2e-7905-4fbc-b941-783049e7835f")
	if err != nil {
		panic(err)
	}
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
