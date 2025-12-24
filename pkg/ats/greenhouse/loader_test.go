package greenhouse

import (
	"context"
	_ "embed"
	"slices"
	"testing"

	"github.com/amalgamated-tools/jobscraping/pkg/ats/models"
	"github.com/amalgamated-tools/jobscraping/pkg/helpers"
)

//go:embed single_job.json
var singleJob string

func Test_parseSingleGreenhouseJob(t *testing.T) {
	t.Parallel()

	job, err := parseGreenhouseJob(context.Background(), []byte(singleJob))
	if err != nil {
		t.Fatalf("parseGreenhouseJob() error = %v", err)
	}

	if job.SourceID != "7454336" {
		t.Errorf("parseGreenhouseJob() SourceID = %v, want %v", job.SourceID, "7454336")
	}

	if job.Title != "Customer Success Manager II, Mid-Market" {
		t.Errorf("parseGreenhouseJob() Title = %v, want %v", job.Title, "Customer Success Manager II, Mid-Market")
	}

	if job.Department != models.SoftwareEngineering {
		t.Errorf("parseGreenhouseJob() Department = %v, want %v", job.Department, "Engineering")
	}
	// if !strings.EqualFold(strings.Join(job.GetMetadata("team"), ","), "EMEA Engineering") {
	// t.Errorf("parseGreenhouseJob() team metadata = %v, want %v", job.GetMetadata("team"), "EMEA Engineering")
	// }
	if job.EmploymentType != models.FullTime {
		t.Errorf("parseGreenhouseJob() EmploymentType = %v, want %v", job.EmploymentType, "FullTime")
	}

	if job.Location != "Remote - Europe" {
		t.Errorf("parseGreenhouseJob() Location = %v, want %v", job.Location, "Remote - Europe")
	}

	if !job.IsRemote {
		t.Errorf("parseGreenhouseJob() IsRemote = %v, want %v", job.IsRemote, true)
	}

	for _, v := range []string{"Barcelona", "Belgium", "France", "Netherlands", "Madrid"} {
		if !slices.Contains(job.GetMetadata("secondary_location"), v) {
			t.Errorf("parseGreenhouseJob() secondary_location metadata missing %v", v)
		}
	}

	if job.DatePosted.IsZero() {
		t.Errorf("parseGreenhouseJob() DatePosted is zero")
	}

	if job.CompensationUnit == nil || helpers.StringValue(job.CompensationUnit) != "€" {
		t.Errorf("parseGreenhouseJob() CompensationUnit = %v, want %v", helpers.StringValue(job.CompensationUnit), "YEAR")
	}

	if job.MinCompensation != 185000 {
		t.Errorf("parseGreenhouseJob() MinCompensation = %v, want %v", job.MinCompensation, 185000)
	}

	if job.MaxCompensation != 317000 {
		t.Errorf("parseGreenhouseJob() MaxCompensation = %v, want %v", job.MaxCompensation, 317000)
	}
}
