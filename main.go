package main

import (
	"context"
	"database/sql"
	_ "embed"

	"github.com/amalgamated-tools/jobscraping/pkg/db"
	_ "modernc.org/sqlite"
)

//go:embed db/schema.sql
var ddl string

func main() {
	ctx := context.Background()

	ab, err := sql.Open("sqlite", "file:jobscraping.db?cache=shared&mode=rwc")
	if err != nil {
		panic(err)
	}
	// create tables
	if _, err := ab.ExecContext(ctx, ddl); err != nil {
		panic(err)
	}

	queries := db.New(ab)
	job, err := queries.CreateJob(ctx, db.CreateJobParams{
		AbsoluteUrl: "https://example.com/job/1",
		Data:        "{ \"source\": \"Example Job\" }",
	})
	if err != nil {
		panic(err)
	}
	_, err = queries.GetJob(ctx, job.ID)
	if err != nil {
		panic(err)
	}
	_, err = queries.ListJobs(ctx)
	if err != nil {
		panic(err)
	}
}
