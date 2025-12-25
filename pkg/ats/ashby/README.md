# Ashby ATS Loader

This package implements a job scraper for the [Ashby ATS (Applicant Tracking System)](https://www.ashbyhq.com/). It provides functions to retrieve and parse job postings from companies using Ashby's hosted job boards.

## Features

- **Company-wide job scraping**: Retrieve all active job postings for a company
- **Individual job scraping**: Fetch detailed information for specific job postings
- **Standardized data model**: Converts Ashby's job data into a common `models.Job` structure
- **Comprehensive parsing**: Extracts job details including:
  - Basic information (title, department, location, employment type)
  - Compensation data (min/max salary, currency, equity)
  - Remote work status (remote, hybrid, onsite)
  - Multiple locations and secondary locations
  - Team information
  - Publication dates
  - Job descriptions (HTML format)

## Installation

This package is part of the `jobscraping` module:

```go
import "github.com/amalgamated-tools/jobscraping/pkg/ats/ashby"
```

## Usage

### Scrape All Jobs for a Company

```go
import (
    "context"
    "github.com/amalgamated-tools/jobscraping/pkg/ats/ashby"
)

func main() {
    ctx := context.Background()
    companyName := "ashby" // The company's Ashby subdomain
    scrapeIndividual := false // Set to true to fetch detailed info for each job
    
    jobs, err := ashby.ScrapeCompany(ctx, companyName, scrapeIndividual)
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
    "github.com/amalgamated-tools/jobscraping/pkg/ats/ashby"
)

func main() {
    ctx := context.Background()
    companyName := "ashby"
    jobID := "6765ef2e-7905-4fbc-b941-783049e7835f"
    
    job, err := ashby.ScrapeJob(ctx, companyName, jobID)
    if err != nil {
        // Handle error
    }
    
    // Process job
    fmt.Printf("Job: %s\n", job.Title)
    fmt.Printf("Compensation: %s%0.0f - %s%0.0f\n", 
        job.CompensationUnit, job.MinCompensation,
        job.CompensationUnit, job.MaxCompensation)
}
```

## API Reference

### Functions

#### `ScrapeCompany(ctx context.Context, companyName string, scrapeIndividual bool) ([]*models.Job, error)`

Scrapes all jobs for a given company from Ashby ATS.

**Parameters:**
- `ctx`: Context for the operation
- `companyName`: The company's Ashby subdomain (e.g., "ashby" for jobs.ashbyhq.com/ashby)
- `scrapeIndividual`: If true, makes additional API calls to fetch detailed information for each job

**Returns:**
- A slice of `*models.Job` containing all job postings
- An error if the scraping fails

#### `ScrapeJob(ctx context.Context, companyName, jobID string) (*models.Job, error)`

Scrapes an individual job from Ashby ATS given the company name and job ID.

**Parameters:**
- `ctx`: Context for the operation
- `companyName`: The company's Ashby subdomain
- `jobID`: The unique identifier for the job posting

**Returns:**
- A `*models.Job` containing the job posting details
- An error if the scraping fails

## Data Format

The package uses two different API endpoints and data formats:

### Company Jobs API

Endpoint: `https://api.ashbyhq.com/posting-api/job-board/{companyName}?includeCompensation=true`

Returns an array of job objects with fields like:
- `id`: Unique job identifier
- `title`: Job title
- `department`: Department name
- `team`: Team name
- `employmentType`: Full-time, part-time, etc.
- `location`: Primary location
- `secondaryLocations`: Array of additional locations
- `isRemote`: Boolean indicating remote status
- `publishedAt`: Publication date
- `jobUrl`: URL to the job posting
- `descriptionHtml`: HTML-formatted job description
- `compensation`: Compensation details including salary range

### Single Job API

Endpoint: `https://jobs.ashbyhq.com/api/non-user-graphql?op=ApiJobPosting`

Uses a GraphQL query to fetch detailed information for a specific job, including:
- `jobPosting`: Nested object with job details
- `locationName`: Primary location name
- `workplaceType`: REMOTE, HYBRID, or ONSITE
- `secondaryLocationNames`: Array of location names
- `teamNames`: Array of team names
- `compensationTierSummary`: Salary range summary (e.g., "$155K - $190K")

## Testing

The package includes comprehensive tests in `loader_test.go`:

```bash
go test ./pkg/ats/ashby/...
```

Tests cover:
- Parsing company job board responses
- Parsing individual job GraphQL responses
- Data extraction and validation
- Compensation parsing
- Location and team metadata handling

Sample JSON files are included:
- `companies_job.json`: Example response from the company jobs API
- `single_job.json`: Example response from the individual job API

## Implementation Details

### Parsing Logic

The package implements two parsing functions:

1. **`parseCompanyJob`**: Parses job data from the company jobs API endpoint
2. **`parseSingleJob`**: Parses job data from the GraphQL API endpoint (nested under `data.jobPosting`)

Both parsers extract data into the same `models.Job` structure for consistency.

### Compensation Parsing

Compensation strings (e.g., "€185K - €317K") are parsed using the `helpers.ParseCompensation` function, which:
- Extracts currency symbols
- Converts string amounts to numeric values
- Handles thousands (K) multipliers
- Detects equity mentions

### Remote Work Detection

The package determines remote work status from:
- `isRemote` boolean (company jobs API)
- `workplaceType` field (single job API) - checks for "REMOTE" value

## Dependencies

- `github.com/amalgamated-tools/jobscraping/pkg/ats/models`: Common job data models
- `github.com/amalgamated-tools/jobscraping/pkg/helpers`: HTTP helpers and parsers
- `github.com/buger/jsonparser`: JSON parsing library

## Contributing

When contributing to this package:

1. Ensure all tests pass: `go test ./pkg/ats/ashby/...`
2. Add tests for new functionality
3. Update sample JSON files if API responses change
4. Follow the existing code style and error handling patterns
5. Update this README if adding new features or changing the API

## Error Handling

The package uses structured logging with `slog` for debugging and error reporting. All public functions return errors that wrap underlying errors with context using `fmt.Errorf` with the `%w` verb.

Common errors:
- Network errors when fetching job data
- JSON parsing errors for malformed responses
- Compensation parsing errors for unrecognized formats
