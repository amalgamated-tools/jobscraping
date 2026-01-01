// Package rippling contains functions to scrape job listings from the Rippling ATS.
package rippling

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/amalgamated-tools/jobscraping/pkg/ats/models"
	"github.com/amalgamated-tools/jobscraping/pkg/helpers"
	"github.com/buger/jsonparser"
)

var (
	ripplingCompanyURL = "https://ats.rippling.com/api/v2/board/%s/jobs"
	ripplingJobURL     = "https://ats.rippling.com/api/v2/board/%s/jobs/%s"
)

// ScrapeCompany scrapes all job listings for a given company from Rippling ATS.
func ScrapeCompany(ctx context.Context, companyName string) ([]*models.Job, error) {
	slog.DebugContext(ctx, "Scraping company", slog.String("ats", "rippling"), slog.String("company_name", companyName))

	jobs := make([]*models.Job, 0)

	// The URL is like https://ats.rippling.com/api/v2/board/%s/jobs
	companyURL := fmt.Sprintf(ripplingCompanyURL, companyName)

	body, err := helpers.GetJSON(ctx, companyURL, nil)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting JSON from Rippling job board endpoint", slog.String("url", companyURL), slog.Any("error", err))
		return jobs, fmt.Errorf("error getting JSON from Rippling job board endpoint: %w", err)
	}

	_, err = jsonparser.ArrayEach(body, func(value []byte, _ jsonparser.ValueType, _ int, _ error) {
		jobID, err := jsonparser.GetString(value, "id")
		if err != nil {
			slog.ErrorContext(ctx, "Error parsing job ID from jobs array", slog.Any("error", err))
			return
		}

		job, jerr := ScrapeJob(ctx, companyName, jobID)
		if jerr != nil {
			slog.ErrorContext(ctx, "Error parsing rippling job from jobs array", slog.Any("error", jerr))
			return
		}

		jobs = append(jobs, job)
	}, "items")
	if err != nil {
		slog.ErrorContext(ctx, "Error parsing jobs array from Rippling job board endpoint", slog.Any("error", err))
		return jobs, fmt.Errorf("error parsing jobs array from Rippling job board endpoint: %w", err)
	}

	return jobs, nil
}

// ScrapeJob scrapes an individual job listing from Rippling ATS.
func ScrapeJob(ctx context.Context, companyName, jobID string) (*models.Job, error) {
	slog.DebugContext(ctx, "Scraping individual job", slog.String("ats", "rippling"), slog.String("company_name", companyName), slog.String("job_id", jobID))

	// The URL is like https://ats.rippling.com/api/v2/board/smartwyre/jobs/698a497a-ab01-48dc-9517-3d25704cc32c
	jobURL := fmt.Sprintf(ripplingJobURL, companyName, jobID)

	body, err := helpers.GetJSON(ctx, jobURL, nil)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting JSON from Rippling job endpoint", slog.String("url", jobURL), slog.Any("error", err))
		return nil, fmt.Errorf("error getting JSON from Rippling job endpoint: %w", err)
	}

	return parseRipplingJob(ctx, body)
}

func parseRipplingJob(ctx context.Context, data []byte) (*models.Job, error) {
	job := models.NewJob("rippling", data)

	err := jsonparser.ObjectEach(job.GetSourceData(), func(key []byte, value []byte, _ jsonparser.ValueType, _ int) error {
		switch string(key) {
		case "uuid":
			job.SourceID = string(value)
		case "name":
			job.Title = string(value)
		case "description":
			role, err := jsonparser.GetString(value, "role")
			if err == nil {
				job.Description = role
			}

			company, err := jsonparser.GetString(value, "company")
			if err == nil {
				job.Company.Description = helpers.Ptr(company)
			}
		case "workLocations":
			_, jerr := jsonparser.ArrayEach(value, func(locValue []byte, _ jsonparser.ValueType, _ int, _ error) {
				location := string(locValue)
				if job.Location == "" {
					job.Location = location
				}

				slog.DebugContext(ctx, "Parsed location", slog.String("location", location))
				job.AddMetadata("workLocations", location)
			})
			if jerr != nil {
				slog.ErrorContext(ctx, "Error parsing workLocations array", slog.Any("error", jerr))
			}
		case "department":
			name, err := jsonparser.GetString(value, "name")
			if err != nil {
				slog.ErrorContext(ctx, "Error parsing department name", slog.Any("error", err))
				return nil
			}

			job.Department = models.ParseDepartment(name)
			job.DepartmentRaw = name
		case "employmentType":
			label, err := jsonparser.GetString(value, "label")
			if err != nil {
				slog.ErrorContext(ctx, "Error parsing employmentType label", slog.Any("error", err))
				return nil
			}

			job.EmploymentType = models.ParseEmploymentType(label)
		case "createdOn":
			job.ProcessDatePosted(ctx, value)
		case "url":
			job.URL = string(value)
		case "board":
			boardURL, err := jsonparser.GetString(value, "boardURL")
			if err == nil {
				homepage, err := url.Parse(boardURL)
				if err == nil {
					job.Company.Homepage = *homepage
				}
			}

			logo, err := jsonparser.GetString(value, "logo")
			if err == nil && logo != "" && logo != "null" {
				logoURL, err := url.Parse(logo)
				if err == nil {
					job.Company.Logo = *logoURL
				}
			}
		case "companyName":
			job.Company.Name = string(value)
		default:
			job.AddMetadata(string(key), string(value))
		}

		return nil
	})
	if err != nil {
		slog.ErrorContext(ctx, "Error parsing Rippling job object", slog.Any("error", err))
		return nil, fmt.Errorf("error parsing Rippling job object: %w", err)
	}

	return job, nil
}
