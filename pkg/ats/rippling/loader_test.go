package rippling

import (
	"context"
	_ "embed"
	"testing"

	"github.com/amalgamated-tools/jobscraping/pkg/ats/models"
)

//go:embed single_job.json
var singleJob string

func Test_parseRipplingJob(t *testing.T) {
	t.Parallel()

	job, err := parseRipplingJob(context.Background(), []byte(singleJob))
	if err != nil {
		t.Fatalf("parseRipplingJob() error = %v", err)
	}

	if job.SourceID != "698a497a-ab01-48dc-9517-3d25704cc32c" {
		t.Errorf("parseRipplingJob() SourceID = %v, want %v", job.SourceID, "698a497a-ab01-48dc-9517-3d25704cc32c")
	}

	if job.Title != "Finance Analyst (Contractor)" {
		t.Errorf("parseRipplingJob() Title = %v, want %v", job.Title, "Finance Analyst (Contractor)")
	}

	if job.Department != models.Unsure {
		t.Errorf("parseRipplingJob() Department = %v, want %v", job.Department, models.Unsure)
	}

	if job.DepartmentRaw != "Finance" {
		t.Errorf("parseRipplingJob() DepartmentRaw = %v, want %v", job.DepartmentRaw, "Finance")
	}

	if job.EmploymentType != models.Contract {
		t.Errorf("parseRipplingJob() EmploymentType = %v, want %v", job.EmploymentType, models.Contract)
	}

	if job.Location != "Remote (South Africa)" {
		t.Errorf("parseRipplingJob() Location = %v, want %v", job.Location, "Remote (South Africa)")
	}

	if job.Company.Name != "Smartwyre" {
		t.Errorf("parseRipplingJob() Company.Name = %v, want %v", job.Company.Name, "Smartwyre")
	}

	if job.Company.HomepageURL == nil || *job.Company.HomepageURL != "https://www.smartwyre.com/careers" {
		t.Errorf("parseRipplingJob() Company.HomepageURL = %v, want %v", job.Company.HomepageURL, "https://www.smartwyre.com/careers")
	}

	if job.URL != "https://ats.rippling.com/smartwyre/jobs/698a497a-ab01-48dc-9517-3d25704cc32c" {
		t.Errorf("parseRipplingJob() URL = %v, want %v", job.URL, "https://ats.rippling.com/smartwyre/jobs/698a497a-ab01-48dc-9517-3d25704cc32c")
	}

	if job.DatePosted.IsZero() {
		t.Errorf("parseRipplingJob() DatePosted is zero")
	}

	// Check metadata for workLocations
	workLocations := job.GetMetadata("workLocations")
	if len(workLocations) != 3 {
		t.Errorf("parseRipplingJob() workLocations count = %v, want %v", len(workLocations), 3)
	}
}
