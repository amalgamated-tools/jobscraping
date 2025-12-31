# Bamboo ATS Package

This package provides functionality to scrape job listings from BambooHR ATS (Applicant Tracking System).

## Overview

BambooHR is a human resources information system (HRIS) that provides an ATS for posting job openings. This package allows you to:

- Scrape all jobs for a company from their BambooHR job board
- Scrape individual job details by job ID
- Parse job data including title, description, compensation, location, and other metadata

## Functions

### ScrapeCompany

```go
func ScrapeCompany(ctx context.Context, companyName string) ([]*models.Job, error)
```

Scrapes all jobs for a given company from BambooHR ATS. The company name is extracted from the subdomain of their BambooHR URL (e.g., `companyname` from `https://companyname.bamboohr.com/careers/list`).

**Parameters:**
- `ctx`: Context for the request
- `companyName`: The subdomain/company identifier in BambooHR

**Returns:**
- A slice of Job pointers containing all jobs found for the company
- An error if the scraping fails

### ScrapeJob

```go
func ScrapeJob(ctx context.Context, companyName, jobID string) (*models.Job, error)
```

Scrapes an individual job from BambooHR ATS given the company name and job ID.

**Parameters:**
- `ctx`: Context for the request
- `companyName`: The subdomain/company identifier in BambooHR
- `jobID`: The unique identifier for the job

**Returns:**
- A Job pointer containing the job details
- An error if the scraping fails

## Job Data Parsed

The package extracts the following information from BambooHR job postings:

- **Source ID**: Job ID from BambooHR
- **Title**: Job title
- **Description**: Full job description (HTML)
- **Department**: Categorized department
- **Employment Type**: Full-time, part-time, etc.
- **Location**: Job location (city, state, country)
- **Location Type**: On-site, remote, or hybrid
- **Compensation**: Salary range if available
- **Date Posted**: When the job was posted
- **URL**: Direct link to the job posting

## Example Usage

```go
import (
    "context"
    "github.com/amalgamated-tools/jobscraping/pkg/ats/bamboo"
)

// Scrape all jobs for a company
jobs, err := bamboo.ScrapeCompany(context.Background(), "beehiiv")
if err != nil {
    // handle error
}

// Scrape a specific job
job, err := bamboo.ScrapeJob(context.Background(), "beehiiv", "25")
if err != nil {
    // handle error
}
```

## API Endpoints

The package uses the following BambooHR API endpoints:

- Job list: `https://{companyName}.bamboohr.com/careers/list`
- Individual job: `https://{companyName}.bamboohr.com/careers/{jobID}/detail`

## Testing

The package includes comprehensive tests with embedded sample JSON responses:

- `job_list.json`: Sample response from the company job list endpoint
- `single_job.json`: Sample response from an individual job endpoint

Run tests with:

```bash
go test ./pkg/ats/bamboo/
```
