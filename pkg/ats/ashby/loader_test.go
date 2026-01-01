package ashby

import (
	"context"
	_ "embed"
	"slices"
	"testing"

	models "github.com/amalgamated-tools/jobscraping/pkg/ats/models"
)

//go:embed single_job.json
var singleJob string

func Test_parseSingleAshbyJob(t *testing.T) {
	t.Parallel()

	job, err := parseAshbyJob(context.Background(), []byte(singleJob))
	if err != nil {
		t.Fatalf("parseAshbyJob() error = %v", err)
	}

	if job.Source != "ashby" {
		t.Errorf("parseAshbyJob() Source = %v, want %v", job.Source, "ashby")
	}

	if job.SourceID != "6765ef2e-7905-4fbc-b941-783049e7835f" {
		t.Errorf("parseAshbyJob() SourceID = %v, want %v", job.SourceID, "6765ef2e-7905-4fbc-b941-783049e7835f")
	}

	if job.Title != "Principal Product Engineer, EU" {
		t.Errorf("parseAshbyJob() Title = %v, want %v", job.Title, "Principal Product Engineer, EU")
	}

	if job.Department != models.SoftwareEngineering {
		t.Errorf("parseAshbyJob() Department = %v, want %v", job.Department, "Engineering")
	}

	if job.DepartmentRaw != "Engineering" {
		t.Errorf("parseAshbyJob() DepartmentRaw = %v, want %v", job.DepartmentRaw, "Engineering")
	}

	if job.EmploymentType != models.FullTime {
		t.Errorf("parseAshbyJob() EmploymentType = %v, want %v", job.EmploymentType, "FullTime")
	}

	if job.Location != "Remote - Europe" {
		t.Errorf("parseAshbyJob() Location = %v, want %v", job.Location, "Remote - Europe")
	}

	if job.LocationType != models.RemoteLocation {
		t.Errorf("parseAshbyJob() LocationType = %v, want %v", job.LocationType, models.RemoteLocation)
	}

	if !job.IsRemote {
		t.Errorf("parseAshbyJob() IsRemote = %v, want %v", job.IsRemote, true)
	}

	for _, v := range []string{"Barcelona", "Belgium", "France", "Netherlands", "Madrid"} {
		if !slices.Contains(job.GetMetadata("secondary_location"), v) {
			t.Errorf("parseAshbyJob() secondary_location metadata missing %v", v)
		}
	}

	if job.DatePosted.IsZero() {
		t.Errorf("parseAshbyJob() DatePosted is zero")
	}

	if job.CompensationUnit != "€" {
		t.Errorf("parseAshbyJob() CompensationUnit = %v, want %v", job.CompensationUnit, "€")
	}

	if job.MinCompensation != 185000 {
		t.Errorf("parseAshbyJob() MinCompensation = %v, want %v", job.MinCompensation, 185000)
	}

	if job.MaxCompensation != 317000 {
		t.Errorf("parseAshbyJob() MaxCompensation = %v, want %v", job.MaxCompensation, 317000)
	}

	if job.Company.Name != "Ashby" {
		t.Errorf("parseAshbyJob() Company.Name = %v, want %v", job.Company.Name, "Ashby")
	}

	if job.Equity != models.EquityOffered {
		t.Errorf("parseAshbyJob() Equity = %v, want %v", job.Equity, models.EquityOffered)
	}

	if job.Company.Homepage.String() != "https://www.ashbyhq.com" {
		t.Errorf("parseAshbyJob() Company.Homepage = %v, want %v", job.Company.Homepage.String(), "https://www.ashbyhq.com")
	}

	if job.Company.Logo.String() != "https://www.ashbyhq.com/logo.png" {
		t.Errorf("parseAshbyJob() Company.Logo = %v, want %v", job.Company.Logo.String(), "https://www.ashbyhq.com/logo.png")
	}
}
