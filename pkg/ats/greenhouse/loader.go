// Package greenhouse implements an ATS loader for Greenhouse ATS.
package greenhouse

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/amalgamated-tools/jobscraping/pkg/ats/models"
	"github.com/amalgamated-tools/jobscraping/pkg/helpers"
	"github.com/buger/jsonparser"
)

var (
	greenhouseCompanyURL = "https://boards-api.greenhouse.io/v1/boards/%s/jobs?content=true&pay_transparency=true"
	greenhouseJobURL     = "https://boards-api.greenhouse.io/v1/boards/%s/jobs/%s?content=true&pay_transparency=true"
)

// ScrapeCompany scrapes all jobs for a given company from Greenhouse ATS.
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
			slog.ErrorContext(ctx, "Error parsing Greenhouse job from jobs array", slog.Any("error", jerr))
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

// ScrapeJob scrapes an individual job from Greenhouse ATS given the company name and job ID.
func ScrapeJob(ctx context.Context, companyName, jobID string) (*models.Job, error) {
	slog.DebugContext(ctx, "Scraping individual job", slog.String("ats", "greenhouse"), slog.String("company_name", companyName), slog.String("job_id", jobID))

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
	job := models.NewJob("greenhouse", data)

	err := jsonparser.ObjectEach(job.GetSourceData(), func(key []byte, value []byte, _ jsonparser.ValueType, _ int) error {
		switch string(key) {
		case "absolute_url":
			job.URL = string(value)
		case "internal_job_id":
			job.AddMetadata("internal_job_id", string(value))
		case "location":
			location, err := jsonparser.GetString(value, "name")
			if err != nil {
				slog.ErrorContext(ctx, "Error parsing location name", slog.Any("error", err))
				return fmt.Errorf("error parsing location name: %w", err)
			}

			job.Location = location
			job.LocationType = models.ParseLocationType(location)

			if job.LocationType == models.UnknownLocationType {
				job.ProcessLocationType([]string{location})
			}
		case "metadata":
			// an array of objects with key and value fields
			_, _ = jsonparser.ArrayEach(value, func(value []byte, _ jsonparser.ValueType, _ int, _ error) {
				metaKey, err := jsonparser.GetString(value, "name")
				if err != nil {
					slog.ErrorContext(ctx, "Error parsing metadata key", slog.Any("error", err))
					// this is not fatal, continue
				}

				metaID, err := jsonparser.GetInt(value, "id")
				if err != nil {
					slog.ErrorContext(ctx, "Error parsing metadata id", slog.Any("error", err))
					// this is not fatal, continue
				}

				// if we don't have a key or id, skip
				if metaKey == "" && metaID == 0 {
					return
				}

				if metaKey == "" {
					metaKey = strconv.FormatInt(metaID, 10)
				}

				metaValue, _, _, err := jsonparser.Get(value, "value")
				if err != nil {
					slog.ErrorContext(ctx, "Error parsing metadata value", slog.Any("error", err))
					return
				}

				job.AddMetadata(metaKey, string(metaValue))
			})
		case "id":
			job.SourceID = string(value)
		case "updated_at":
			job.AddMetadata("updated_at", string(value))
		case "requisition_id":
			job.AddMetadata("requisition_id", string(value))
		case "title":
			job.Title = string(value)
		case "pay_input_ranges":
			mixCents, err := jsonparser.GetInt(value, "[0]", "min_cents")
			if err == nil {
				job.MinCompensation = float64(mixCents)
			}

			maxCents, err := jsonparser.GetInt(value, "[0]", "max_cents")
			if err == nil {
				job.MaxCompensation = float64(maxCents)
			}

			currencyType, err := jsonparser.GetString(value, "[0]", "currency_type")
			if err == nil {
				job.CompensationUnit = currencyType
			}
		case "company_name":
			job.AddMetadata("company_name", string(value))
		case "first_published":
			job.ProcessDatePosted(ctx, value)
		case "language":
			job.AddMetadata("language", string(value))
		case "content":
			job.Description = string(value)
		case "departments":
			_, _ = jsonparser.ArrayEach(value, func(deptValue []byte, _ jsonparser.ValueType, _ int, _ error) {
				// get the name
				deptName, err := jsonparser.GetString(deptValue, "name")
				if err != nil {
					slog.ErrorContext(ctx, "Error getting department name", slog.Any("error", err))
					return
				}

				if job.DepartmentRaw == "" {
					// we haven't set it yet
					job.DepartmentRaw = deptName
					job.Department = models.ParseDepartment(deptName)
				} else if job.Department == models.UnknownDepartment {
					job.Department = models.ParseDepartment(deptName)
				}

				job.AddMetadata("department", deptName)
			})
		case "offices":
			job.AddMetadata("offices", string(value))
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
