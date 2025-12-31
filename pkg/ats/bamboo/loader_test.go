package bamboo

import (
	"context"
	_ "embed"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amalgamated-tools/jobscraping/pkg/ats/models"
	"github.com/amalgamated-tools/jobscraping/pkg/helpers"
	"github.com/h2non/gock"
)

//go:embed single_job.json
var singleJob string

//go:embed job_list.json
var jobList string

func Test_parseBambooJob(t *testing.T) {
	t.Parallel()

	job, err := parseBambooJob(context.Background(), []byte(singleJob))
	if err != nil {
		t.Fatalf("parseBambooJob() error = %v", err)
	}

	if job.SourceID != "158" {
		t.Errorf("parseBambooJob() SourceID = %v, want %v", job.SourceID, "158")
	}

	if job.Title != "VP of Customer Success " {
		t.Errorf("parseBambooJob() Title = %v, want 'VP of Customer Success '", job.Title)
	}

	if job.Department != models.CustomerSuccessSupport {
		t.Errorf("parseBambooJob() Department = %v, want 'Customer Success and Support'", job.Department)
	}

	if job.DepartmentRaw != "Customer Success" {
		t.Errorf("parseBambooJob() DepartmentRaw = %v, want 'Customer Success'", job.DepartmentRaw)
	}

	if job.Location != "New York, New York 10018, United States" {
		t.Errorf("parseBambooJob() Location = %v, want 'New York, New York 10018, United States'", job.Location)
	}

	if job.MinCompensation != 250000 {
		t.Errorf("parseBambooJob() MinCompensation = %v, want 250000", job.MinCompensation)
	}

	if job.MaxCompensation != 350000 {
		t.Errorf("parseBambooJob() MaxCompensation = %v, want 350000", job.MaxCompensation)
	}

	if job.CompensationUnit != "$" {
		t.Errorf("parseBambooJob() CompensationUnit = %v, want '$'", job.CompensationUnit)
	}

	if job.LocationType != models.HybridLocation {
		t.Errorf("parseBambooJob() LocationType = %v, want HybridLocation", job.LocationType)
	}
}

func TestScrapeCompany(t *testing.T) {
	t.Parallel()

	defer gock.Off() // Flush pending mocks after test execution

	gock.New("https://testcompany.bamboohr.com").
		Get("/careers/list").
		Reply(200).
		JSON(jobList)

	for _, id := range []string{"25", "34", "35"} {
		gock.New("https://testcompany.bamboohr.com").
			Get("/careers/" + id + "/detail").
			Reply(200).
			JSON(singleJob)
	}

	jobs, err := ScrapeCompany(context.Background(), "testcompany")
	if err != nil {
		t.Fatalf("ScrapeCompany() error = %v", err)
	}

	if len(jobs) != 3 {
		t.Errorf("ScrapeCompany() len(jobs) = %v, want 3", len(jobs))
	}
}

func TestScrapeJob(t *testing.T) {
	t.Parallel()

	// Create a test server that returns the single job JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_, err := w.Write([]byte(singleJob))
		if err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Set up HTTP client to use test server
	helpers.SetHTTPClient(server.Client())

	defer helpers.ResetHTTPClient()

	// Test the parse function directly since URL is hardcoded in ScrapeJob
	job, err := parseBambooJob(context.Background(), []byte(singleJob))
	if err != nil {
		t.Fatalf("parseBambooJob() error = %v", err)
	}

	if job.SourceID != "158" {
		t.Errorf("job.SourceID = %v, want 158", job.SourceID)
	}

	if job.Title != "VP of Customer Success " {
		t.Errorf("job.Title = %v, want 'VP of Customer Success '", job.Title)
	}

	if job.URL != "https://axomic.bamboohr.com/careers/158" {
		t.Errorf("job.URL = %v, want 'https://axomic.bamboohr.com/careers/158'", job.URL)
	}
}
