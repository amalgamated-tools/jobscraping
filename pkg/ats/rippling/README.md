# Rippling ATS Loader

This package implements a job scraper for the [Rippling ATS (Applicant Tracking System)](https://www.rippling.com/). It provides functions to retrieve and parse job postings from companies using Rippling's hosted job boards.

## Features

- **Company-wide job scraping**: Retrieve all active job postings for a company
- **Individual job scraping**: Fetch detailed information for specific job postings
- **Standardized data model**: Converts Rippling's job data into a common `models.Job` structure
- **Comprehensive parsing**: Extracts job details including:
  - Basic information (title, department, location, employment type)
  - Work locations (including multiple remote locations)
  - Company information (name, homepage URL, description)
  - Publication dates
  - Job descriptions (HTML format)

## Installation

This package is part of the `jobscraping` module:

```go
import "github.com/amalgamated-tools/jobscraping/pkg/ats/rippling"
```

## Usage

### Scrape All Jobs for a Company

```go
import (
    "context"
    "github.com/amalgamated-tools/jobscraping/pkg/ats/rippling"
)

func main() {
    ctx := context.Background()
    companyName := "smartwyre" // The company's Rippling subdomain
    
    jobs, err := rippling.ScrapeCompany(ctx, companyName)
    if err != nil {
        // Handle error
    }
    
    // Process jobs
    for _, job := range jobs {
        fmt.Printf("Job: %s - %s\n", job.Title, job.Location)
    }
}
```

### Scrape a Single Job

```go
import (
    "context"
    "github.com/amalgamated-tools/jobscraping/pkg/ats/rippling"
)

func main() {
    ctx := context.Background()
    companyName := "smartwyre"
    jobID := "698a497a-ab01-48dc-9517-3d25704cc32c"
    
    job, err := rippling.ScrapeJob(ctx, companyName, jobID)
    if err != nil {
        // Handle error
    }
    
    // Process job
    fmt.Printf("Job: %s\n", job.Title)
    fmt.Printf("Location: %s\n", job.Location)
    fmt.Printf("Company: %s\n", job.Company.Name)
}
```

## API Reference

### Functions

#### `ScrapeCompany(ctx context.Context, companyName string) ([]*models.Job, error)`

Scrapes all jobs for a given company from Rippling ATS.

**Parameters:**
- `ctx`: Context for the operation
- `companyName`: The company's Rippling subdomain (e.g., "smartwyre" for ats.rippling.com/smartwyre)

**Returns:**
- A slice of `*models.Job` containing all job postings
- An error if the scraping fails

**Note:** This function makes API calls to fetch detailed information for each job listing.

#### `ScrapeJob(ctx context.Context, companyName, jobID string) (*models.Job, error)`

Scrapes an individual job from Rippling ATS given the company name and job ID.

**Parameters:**
- `ctx`: Context for the operation
- `companyName`: The company's Rippling subdomain
- `jobID`: The unique identifier for the job posting

**Returns:**
- A `*models.Job` containing the job posting details
- An error if the scraping fails

## Data Format

The package uses two different API endpoints:

### Job List API

Endpoint: `https://ats.rippling.com/api/v2/board/{companyName}/jobs`

Returns a paginated list of jobs with fields like:
- `items`: Array of job objects
- `id`: Unique job identifier
- `name`: Job title
- `url`: URL to the job posting
- `department`: Department information
- `locations`: Array of location objects

### Single Job API

Endpoint: `https://ats.rippling.com/api/v2/board/{companyName}/jobs/{jobID}`

Returns detailed information for a specific job, including:
- `uuid`: Unique job identifier
- `name`: Job title
- `description`: Object containing:
  - `role`: Job description HTML
  - `company`: Company description HTML
- `workLocations`: Array of location strings
- `department`: Department details including name and base_department
- `employmentType`: Employment type with label and id
- `createdOn`: Publication date
- `board`: Job board information including company website URL
- `companyName`: Company name
- `url`: URL to the job posting

## Testing

The package includes comprehensive tests in `loader_test.go`:

```bash
go test ./pkg/ats/rippling/...
```

Tests cover:
- Parsing individual job responses
- Data extraction and validation
- Department and employment type parsing
- Location and metadata handling

Sample JSON files are included:
- `job_list.json`: Example response from the job list API
- `single_job.json`: Example response from the individual job API

## Implementation Details

### Parsing Logic

The package implements a single parsing function:

- **`parseRipplingJob`**: Parses job data from the single job API endpoint

The parser extracts data into the `models.Job` structure, including:
- Job metadata (title, ID, URL, dates)
- Company information
- Department classification
- Employment type
- Work locations
- Job descriptions

### Department Parsing

Department information is extracted from the `department.name` field and classified using `models.ParseDepartment`.

### Employment Type Detection

Employment types are parsed from the `employmentType.label` field and mapped to standard types like:
- Full-time
- Part-time
- Contractor
- Internship

### Location Handling

The package handles multiple work locations:
- Primary location is set from the first item in `workLocations`
- All locations are stored as metadata under the "workLocations" key

## Dependencies

- `github.com/amalgamated-tools/jobscraping/pkg/ats/models`: Common job data models
- `github.com/amalgamated-tools/jobscraping/pkg/helpers`: HTTP helpers and parsers
- `github.com/buger/jsonparser`: JSON parsing library

## Contributing

When contributing to this package:

1. Ensure all tests pass: `go test ./pkg/ats/rippling/...`
2. Add tests for new functionality
3. Update sample JSON files if API responses change
4. Follow the existing code style and error handling patterns
5. Update this README if adding new features or changing the API

## Error Handling

The package uses structured logging with `slog` for debugging and error reporting. All public functions return errors that wrap underlying errors with context using `fmt.Errorf` with the `%w` verb.

Common errors:
- Network errors when fetching job data
- JSON parsing errors for malformed responses
- HTTP non-200 status codes
