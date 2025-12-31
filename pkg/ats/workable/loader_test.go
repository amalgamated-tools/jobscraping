package workable

import (
	_ "embed"
	"testing"
)

//go:embed single_job.json
var singleJob string

func Test_parseWorkableJob(t *testing.T) {
	t.Parallel()

	job, err := parseWorkableJob(t.Context(), []byte(singleJob))
	if err != nil {
		t.Fatalf("parseWorkableJob() error = %v", err)
	}

	if job.SourceID != "741A8A4254" {
		t.Errorf("parseWorkableJob() SourceID = %v, want %v", job.SourceID, "741A8A4254")
	}
}
