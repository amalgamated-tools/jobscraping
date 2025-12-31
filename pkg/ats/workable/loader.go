// Package workable provides functions to scrape job postings from the Workable ATS.
package workable

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/amalgamated-tools/jobscraping/pkg/ats/models"
	"github.com/amalgamated-tools/jobscraping/pkg/helpers"
	"github.com/buger/jsonparser"
)

var (
	workableCompanyURL = "https://apply.workable.com/api/v3/accounts/%s/jobs"
	workableJobURL     = "https://apply.workable.com/api/v2/accounts/%s/jobs/%s"
)

// ScrapeCompany scrapes all jobs for a given company from Workable ATS.
func ScrapeCompany(ctx context.Context, companyName string) ([]*models.Job, error) {
	slog.DebugContext(ctx, "Scraping company", slog.String("ats", "workable"), slog.String("company_name", companyName))

	jobs := make([]*models.Job, 0)

	companyURL := fmt.Sprintf(workableCompanyURL, companyName)

	payload := strings.NewReader(`{"query":"","department":[],"location":[],"remote":[],"workplace":[],"worktype":[]}`)

	body, err := helpers.PostJSON(ctx, companyURL, payload, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch jobs for company %s: %w", companyName, err)
	}

	_, err = jsonparser.ArrayEach(body, func(value []byte, _ jsonparser.ValueType, _ int, _ error) {
		shortcode, err := jsonparser.GetString(value, "shortcode")
		if err != nil {
			slog.ErrorContext(ctx, "Failed to get job shortcode", slog.String("ats", "workable"), slog.String("company_name", companyName), slog.Any("error", err))
			return
		}

		job, err := ScrapeJob(ctx, companyName, shortcode)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to parse job", slog.String("ats", "workable"), slog.String("company_name", companyName), slog.Any("error", err))
			return
		}

		jobs = append(jobs, job)
	}, "jobs")
	if err != nil {
		return nil, fmt.Errorf("failed to parse jobs for company %s: %w", companyName, err)
	}

	return jobs, nil
}

// ScrapeJob scrapes an individual job from Workable ATS given the company name and job ID.
func ScrapeJob(ctx context.Context, companyName, jobID string) (*models.Job, error) {
	slog.DebugContext(ctx, "Scraping job", slog.String("ats", "workable"), slog.String("company_name", companyName), slog.String("job_id", jobID))
	url := fmt.Sprintf(workableJobURL, companyName, jobID)

	body, err := helpers.GetJSON(ctx, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch job %s for company %s: %w", jobID, companyName, err)
	}

	return parseWorkableJob(ctx, body)
}

func parseWorkableJob(ctx context.Context, data []byte) (*models.Job, error) {
	slog.DebugContext(ctx, "Parsing Workable job data", slog.String("ats", "workable"))

	job := models.NewJob("workable", data)

	err := jsonparser.ObjectEach(job.GetSourceData(), func(key []byte, value []byte, _ jsonparser.ValueType, _ int) error {
		switch string(key) {
		case "id":
			id, err := jsonparser.GetString(value)
			if err == nil {
				job.AddMetadata("workable_id", id)
			}
		case "shortcode":
			// this is definitely a string but jsonparser thinks it is a number
			shortcode, _, _, err := jsonparser.Get(value)
			if err != nil {
				slog.ErrorContext(ctx, "Failed to parse shortcode", slog.String("ats", "workable"), slog.Any("error", err))
				return fmt.Errorf("error parsing shortcode: %w", err)
			}

			job.SourceID = string(shortcode)
		case "title":
			job.Title = string(value)
		case "remote":
			remote, err := jsonparser.GetBoolean(value)
			if err == nil {
				job.IsRemote = remote
			}
		case "location":
			location := models.ParseLocation(value)
			if job.Location == "" {
				job.Location = location.String()
			}

			job.AddMetadata("parsed_location", location.String())
		case "locations":
			_, err := jsonparser.ArrayEach(value, func(locValue []byte, _ jsonparser.ValueType, _ int, _ error) {
				location := models.ParseLocation(locValue)
				if job.Location == "" {
					job.Location = location.String()
				}

				job.AddMetadata("parsed_location", location.String())
			})
			if err != nil {
				slog.ErrorContext(ctx, "Failed to parse locations", slog.String("ats", "workable"), slog.Any("error", err))
				return fmt.Errorf("error parsing locations: %w", err)
			}
		case "published":
			job.ProcessDatePosted(ctx, value)
		case "type":
			job.EmploymentType = models.ParseEmploymentType(string(value))
		case "department":
			_, err := jsonparser.ArrayEach(value, func(deptValue []byte, _ jsonparser.ValueType, _ int, _ error) {
				department, err := jsonparser.GetString(deptValue)
				if err != nil {
					slog.ErrorContext(ctx, "Failed to parse department", slog.String("ats", "workable"), slog.Any("error", err))
					return
				}

				if job.Department == models.UnknownDepartment {
					job.Department = models.ParseDepartment(department)
				}

				job.AddMetadata("parsed_department", department)
			})
			if err != nil {
				slog.ErrorContext(ctx, "Failed to parse departments", slog.String("ats", "workable"), slog.Any("error", err))
				return fmt.Errorf("error parsing departments: %w", err)
			}
		case "workplace":
			job.LocationType = models.ParseLocationType(string(value))
		case "description":
			job.Description = string(value)
		default:
			job.AddMetadata(string(key), string(value))
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error parsing Workable job object: %w", err)
	}

	return job, nil
}
