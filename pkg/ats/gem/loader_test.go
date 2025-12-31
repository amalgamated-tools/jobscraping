package gem

import (
	_ "embed"
	"testing"

	"github.com/amalgamated-tools/jobscraping/pkg/ats/models"
)

//go:embed company_job.json
var companyJob string

//go:embed single_job.json
var singleJob string

func Test_parseGemCompanyJob(t *testing.T) {
	t.Parallel()

	job, err := parseGemCompanyJob(t.Context(), []byte(companyJob))
	if err != nil {
		t.Fatalf("parseGemCompanyJob() error = %v", err)
	}

	if job.SourceID != "am9iOlINaThFqbrhpjCKIMpLj9E" {
		t.Errorf("parseGemCompanyJob() SourceID = %v, want %v", job.SourceID, "am9iOlINaThFqbrhpjCKIMpLj9E")
	}

	if job.URL != "https://jobs.gem.com/arlo/am9icG9zdDruRXVwItfMiCqE7gmjrD4Q" {
		t.Errorf("parseGemCompanyJob() URL = %v, want %v", job.URL, "https://jobs.gem.com/arlo/am9icG9zdDruRXVwItfMiCqE7gmjrD4Q")
	}

	if job.DatePosted.IsZero() {
		t.Errorf("parseGemCompanyJob() DatePosted is zero")
	}

	if job.Department != models.UnknownDepartment {
		t.Errorf("parseGemCompanyJob() Department = %v, want %v", job.Department, models.SoftwareEngineering)
	}

	if job.DepartmentRaw != "Engineering & Data Science" {
		t.Errorf("parseGemCompanyJob() DepartmentRaw = %v, want %v", job.DepartmentRaw, "Engineering & Data Science")
	}

	if job.EmploymentType != models.FullTime {
		t.Errorf("parseGemCompanyJob() EmploymentType = %v, want %v", job.EmploymentType, models.FullTime)
	}

	if job.LocationType != models.OnsiteLocation {
		t.Errorf("parseGemCompanyJob() LocationType = %v, want %v", job.LocationType, models.OnsiteLocation)
	}

	if job.Title != "Founding Software Engineer | Data Platform" {
		t.Errorf("parseGemCompanyJob() Title = %v, want %v", job.Title, "Founding Software Engineer | Data Platform")
	}
}

func Test_parseGemOatsJob(t *testing.T) {
	t.Parallel()

	job, err := parseGemOatsJob(t.Context(), []byte(singleJob))
	if err != nil {
		t.Fatalf("parseGemOatsJob() error = %v", err)
	}

	if job.SourceID != "am9icG9zdDruRXVwItfMiCqE7gmjrD4Q" {
		t.Errorf("parseGemCompanyJob() SourceID = %v, want %v", job.SourceID, "am9icG9zdDruRXVwItfMiCqE7gmjrD4Q")
	}

	if job.DatePosted.IsZero() {
		t.Errorf("parseGemCompanyJob() DatePosted is zero")
	}

	if job.Department != models.UnknownDepartment {
		t.Errorf("parseGemCompanyJob() Department = %v, want %v", job.Department, models.SoftwareEngineering)
	}

	if job.DepartmentRaw != "Engineering & Data Science" {
		t.Errorf("parseGemCompanyJob() DepartmentRaw = %v, want %v", job.DepartmentRaw, "Engineering & Data Science")
	}

	if job.EmploymentType != models.FullTime {
		t.Errorf("parseGemCompanyJob() EmploymentType = %v, want %v", job.EmploymentType, models.FullTime)
	}

	if job.LocationType != models.OnsiteLocation {
		t.Errorf("parseGemCompanyJob() LocationType = %v, want %v", job.LocationType, models.OnsiteLocation)
	}

	if job.Title != "Founding Software Engineer | Data Platform" {
		t.Errorf("parseGemCompanyJob() Title = %v, want %v", job.Title, "Founding Software Engineer | Data Platform")
	}
}
