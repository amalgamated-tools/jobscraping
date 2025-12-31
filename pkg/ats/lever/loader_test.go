package lever

import (
	"context"
	_ "embed"
	"testing"

	"github.com/amalgamated-tools/jobscraping/pkg/ats/models"
)

//go:embed single_job.json
var singleJob string

func Test_parseLeverJob(t *testing.T) {
	t.Parallel()

	job, err := parseLeverJob(context.Background(), []byte(singleJob))
	if err != nil {
		t.Fatalf("parseLeverJob() error = %v", err)
	}

	if job.SourceID != "e002c7c5-c91d-46d0-b23e-62bccdb1695c" {
		t.Errorf("parseLeverJob() SourceID = %v, want %v", job.SourceID, "e002c7c5-c91d-46d0-b23e-62bccdb1695c")
	}

	if job.Title != "Corporate Sales Manager, France (h/f/x)" {
		t.Errorf("parseLeverJob() Title = %v, want %v", job.Title, "Corporate Sales Manager, France (h/f/x)")
	}

	if job.EmploymentType != models.FullTime {
		t.Errorf("parseLeverJob() EmploymentType = %v, want %v", job.EmploymentType, "FullTime")
	}

	if job.Location != "France" {
		t.Errorf("parseLeverJob() Location = %v, want %v", job.Location, "France")
	}

	if !job.IsRemote {
		t.Errorf("parseLeverJob() IsRemote = %v, want %v", job.IsRemote, true)
	}

	if job.DatePosted.IsZero() {
		t.Errorf("parseLeverJob() DatePosted is zero")
	}

	if job.MinCompensation != 187000 {
		t.Errorf("parseLeverJob() MinCompensation = %v, want %v", job.MinCompensation, 187000)
	}

	if job.MaxCompensation != 245000 {
		t.Errorf("parseLeverJob() MaxCompensation = %v, want %v", job.MaxCompensation, 245000)
	}

	if job.LocationType != models.RemoteLocation {
		t.Errorf("parseLeverJob() LocationType = %v, want %v", job.LocationType, "RemoteLocation")
	}

	if job.Department != models.UnknownDepartment {
		t.Errorf("parseLeverJob() Department = %v, want %v", job.Department, "Unsure")
	}
}
