// Package greenhouse implements an ATS loader for Greenhouse ATS.
package greenhouse

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/amalgamated-tools/jobscraping/pkg/ats/models"
	"github.com/amalgamated-tools/jobscraping/pkg/helpers"
	"github.com/buger/jsonparser"
)

var (
	greenhouseCompanyURL = "https://boards-api.greenhouse.io/v1/boards/%s/jobs?content=true&pay_transparency=true"
	greenhouseJobURL     = "https://boards-api.greenhouse.io/v1/boards/%s/jobs/%d?content=true&pay_transparency=true"
)

// ScrapeCompany scrapes all jobs for a given company from Ashby ATS.
func ScrapeCompany(ctx context.Context, companyName string) ([]*models.Job, error) {
	slog.DebugContext(ctx, "Scraping company", slog.String("ats", "greenhouse"), slog.String("company_name", companyName))

	jobs := make([]*models.Job, 0)

	// The URL is like https://boards-api.greenhouse.io/v1/boards/{companyName}/jobs?content=true
	companyURL := fmt.Sprintf(greenhouseCompanyURL, companyName)

	// Get the JSON from the company job board endpoint
	body, err := helpers.GetJSON(ctx, companyURL, nil)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting JSON from Greenhouse job board endpoint", slog.String("url", companyURL), slog.Any("error", err))
		return jobs, fmt.Errorf("error getting JSON from Greenhouse job board endpoint: %w", err)
	}

	_, err = jsonparser.ArrayEach(body, func(value []byte, _ jsonparser.ValueType, _ int, _ error) {
		job, jerr := parseGreenhouseJob(ctx, value)
		if jerr != nil {
			slog.ErrorContext(ctx, "Error parsing Ashby job from jobs array", slog.Any("error", jerr))
			return
		}

		slog.DebugContext(ctx, "Parsed job", slog.String("job_id", job.SourceID), slog.String("title", job.Title))
		jobs = append(jobs, job)
	}, "jobs")
	if err != nil {
		slog.ErrorContext(ctx, "Error parsing jobs array from Greenhouse job board endpoint", slog.Any("error", err))
		return jobs, fmt.Errorf("error parsing jobs array: %w", err)
	}

	return jobs, nil
}

// ScrapeJob scrapes an individual job from Ashby ATS given the company name and job ID.
func ScrapeJob(ctx context.Context, companyName, jobID string) (*models.Job, error) {
	slog.DebugContext(ctx, "Scraping individual job", slog.String("ats", "ashby"), slog.String("company_name", companyName), slog.String("job_id", jobID))

	// The URL is like https://boards-api.greenhouse.io/v1/boards/{companyName}/jobs/{jobID}?content=true
	jobURL := fmt.Sprintf(greenhouseJobURL, companyName, jobID)

	// Get the JSON from the job endpoint
	body, err := helpers.GetJSON(ctx, jobURL, nil)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting JSON from Greenhouse job endpoint", slog.String("url", jobURL), slog.Any("error", err))
		return nil, fmt.Errorf("error getting JSON from Greenhouse job endpoint: %w", err)
	}

	job, err := parseGreenhouseJob(ctx, body)
	if err != nil {
		slog.ErrorContext(ctx, "Error parsing Greenhouse job from job endpoint", slog.Any("error", err))
		return nil, fmt.Errorf("error parsing Greenhouse job from job endpoint: %w", err)
	}

	slog.DebugContext(ctx, "Parsed job", slog.String("job_id", job.SourceID), slog.String("title", job.Title))

	return job, nil
}

func parseGreenhouseJob(ctx context.Context, data []byte) (*models.Job, error) {
	job := &models.Job{
		Source: "greenhouse",
	}
	job.SetSourceData(data)

	err := jsonparser.ObjectEach(job.GetSourceData(), func(key []byte, value []byte, _ jsonparser.ValueType, _ int) error {
		switch string(key) {
		case "absolute_url":
			job.URL = string(value)
		case "internal_job_id":
			job.AddMetadata("internal_job_id", string(value))
		case "id":
			job.SourceID = string(value)
		case "location":
			location, err := jsonparser.GetString(value, "name")
			if err != nil {
				slog.ErrorContext(ctx, "Error parsing location name", slog.Any("error", err))
				return fmt.Errorf("error parsing location name: %w", err)
			}

			job.Location = location
		case "requisition_id":
			job.AddMetadata("requisition_id", string(value))
		case "title":
			job.Title = string(value)
		case "pay_input_ranges":
			// An array of objects, each containing min_cents, max_cents, currency_type, title, and blurb.
			// "minCompensationInCents": {"pay_input_ranges", "[0]", "min_cents"},
			// "maxCompensationInCents": {"pay_input_ranges", "[0]", "max_cents"},
			// "compensationUnit":       {"pay_input_ranges", "[0]", "currency_type"},
			// "payBlurb":               {"pay_input_ranges", "[0]", "blurb"},
		case "company_name":
			job.AddMetadata("company_name", string(value))
		case "first_published":
			job.ProcessDatePosted(ctx, value)
		case "departments":
			job.Department = models.ParseDepartment(string(value))
		case "offices":
		case "content":
			job.Description = string(value)
		default:
			job.AddMetadata(string(key), string(value))
		}

		return nil
	})
	if err != nil {
		slog.ErrorContext(ctx, "Error parsing job object", slog.Any("error", err))
		return job, fmt.Errorf("error parsing job object: %w", err)
	}

	// return nil
	return job, nil
}
