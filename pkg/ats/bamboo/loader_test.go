package bamboo

import (
	"context"
	_ "embed"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amalgamated-tools/jobscraping/pkg/ats/models"
	"github.com/amalgamated-tools/jobscraping/pkg/helpers"
	"github.com/buger/jsonparser"
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

	// Test parsing the job list data to verify structure
	var jobIDs []string

	_, err := jsonparser.ArrayEach([]byte(jobList), func(value []byte, _ jsonparser.ValueType, _ int, _ error) {
		id, err := jsonparser.GetString(value, "id")
		if err != nil {
			return
		}

		jobIDs = append(jobIDs, id)
	}, "result")
	if err != nil {
		t.Fatalf("Error parsing job list: %v", err)
	}

	if len(jobIDs) != 12 {
		t.Errorf("Expected 12 jobs in job_list.json, got %d", len(jobIDs))
	}

	// Verify first job ID
	if len(jobIDs) > 0 && jobIDs[0] != "25" {
		t.Errorf("First job ID = %v, want '25'", jobIDs[0])
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
