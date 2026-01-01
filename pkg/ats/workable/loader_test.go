package workable

import (
	_ "embed"
	"testing"

	"github.com/amalgamated-tools/jobscraping/pkg/ats/models"
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

	if job.Title != "Event & Webinar Coordinator (Part-Time)" {
		t.Errorf("parseWorkableJob() Title = %v, want %v", job.Title, "Event & Webinar Coordinator (Part-Time)")
	}

	if job.IsRemote {
		t.Errorf("parseWorkableJob() IsRemote = %v, want %v", job.IsRemote, false)
	}

	if job.Location != "Mexico City, Mexico City, Mexico" {
		t.Errorf("parseWorkableJob() Location = %v, want %v", job.Location, "Mexico City, Mexico City, Mexico")
	}

	if job.LocationType != models.HybridLocation {
		t.Errorf("parseWorkableJob() LocationType = %v, want %v", job.LocationType, models.HybridLocation)
	}

	if job.DatePosted.IsZero() {
		t.Errorf("parseWorkableJob() DatePosted is zero, want non-zero value")
	}

	if job.EmploymentType != models.PartTime {
		t.Errorf("parseWorkableJob() EmploymentType = %v, want %v", job.EmploymentType, "part_time")
	}

	if job.Department != models.Marketing {
		t.Errorf("parseWorkableJob() Department = %v, want %v", job.Department.String(), "marketing")
	}

	if job.DepartmentRaw != "Marketing" {
		t.Errorf("parseWorkableJob() DepartmentRaw = %v, want %v", job.DepartmentRaw, "Marketing")
	}
}
