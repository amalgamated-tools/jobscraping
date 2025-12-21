package ashby

import (
	_ "embed"
	"slices"
	"strings"
	"testing"

	"github.com/amalgamated-tools/jobscraping/pkg/helpers"
	"github.com/amalgamated-tools/jobscraping/pkg/models"
)

//go:embed companies_job.json
var companies_job string

//go:embed single_job.json
var single_job string

func Test_parseAshbyJob(t *testing.T) {
	job, err := parseAshbyJob([]byte(companies_job))
	if err != nil {
		t.Fatalf("parseAshbyJob() error = %v", err)
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
	if !strings.EqualFold(strings.Join(job.GetMetadata("team"), ","), "EMEA Engineering") {
		t.Errorf("parseAshbyJob() team metadata = %v, want %v", job.GetMetadata("team"), "EMEA Engineering")
	}
	if job.EmploymentType != models.FullTime {
		t.Errorf("parseAshbyJob() EmploymentType = %v, want %v", job.EmploymentType, "FullTime")
	}
	if job.Location != "Remote - Europe" {
		t.Errorf("parseAshbyJob() Location = %v, want %v", job.Location, "Remote - Europe")
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
	if job.URL != "https://jobs.ashbyhq.com/ashby/6765ef2e-7905-4fbc-b941-783049e7835f" {
		t.Errorf("parseAshbyJob() URL = %v, want %v", job.URL, "https://jobs.ashbyhq.com/ashby/6765ef2e-7905-4fbc-b941-783049e7835f")
	}
	if job.CompensationUnit == nil || helpers.StringValue(job.CompensationUnit) != "€" {
		t.Errorf("parseAshbyJob() CompensationUnit = %v, want %v", helpers.StringValue(job.CompensationUnit), "YEAR")
	}
	if job.MinCompensation != 185000 {
		t.Errorf("parseAshbyJob() MinCompensation = %v, want %v", job.MinCompensation, 185000)
	}
	if job.MaxCompensation != 317000 {
		t.Errorf("parseAshbyJob() MaxCompensation = %v, want %v", job.MaxCompensation, 317000)
	}
}

func Test_parseSingleAshbyJob(t *testing.T) {
	job, err := parseAshbyJob([]byte(single_job))
	if err != nil {
		t.Fatalf("parseAshbyJob() error = %v", err)
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
	// if !strings.EqualFold(strings.Join(job.GetMetadata("team"), ","), "EMEA Engineering") {
	// t.Errorf("parseAshbyJob() team metadata = %v, want %v", job.GetMetadata("team"), "EMEA Engineering")
	// }
	if job.EmploymentType != models.FullTime {
		t.Errorf("parseAshbyJob() EmploymentType = %v, want %v", job.EmploymentType, "FullTime")
	}
	if job.Location != "Remote - Europe" {
		t.Errorf("parseAshbyJob() Location = %v, want %v", job.Location, "Remote - Europe")
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

	if job.CompensationUnit == nil || helpers.StringValue(job.CompensationUnit) != "€" {
		t.Errorf("parseAshbyJob() CompensationUnit = %v, want %v", helpers.StringValue(job.CompensationUnit), "YEAR")
	}
	if job.MinCompensation != 185000 {
		t.Errorf("parseAshbyJob() MinCompensation = %v, want %v", job.MinCompensation, 185000)
	}
	if job.MaxCompensation != 317000 {
		t.Errorf("parseAshbyJob() MaxCompensation = %v, want %v", job.MaxCompensation, 317000)
	}
}
