// Package bamboo contains functions to scrape jobs from BambooHR ATS.
package bamboo

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/amalgamated-tools/jobscraping/pkg/ats/models"
	"github.com/amalgamated-tools/jobscraping/pkg/helpers"
	"github.com/buger/jsonparser"
)

// ScrapeCompany scrapes all jobs for a given company from BambooHR ATS.
func ScrapeCompany(ctx context.Context, companyName string) ([]*models.Job, error) {
	slog.DebugContext(ctx, "Scraping company", slog.String("ats", "bamboo"), slog.String("company_name", companyName))

	jobs := make([]*models.Job, 0)

	companyURL := "https://" + companyName + ".bamboohr.com/careers/list"

	body, err := helpers.GetJSON(ctx, companyURL, nil)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting JSON from BambooHR job board endpoint", slog.String("url", companyURL), slog.Any("error", err))
		return jobs, fmt.Errorf("error getting JSON from BambooHR job board endpoint: %w", err)
	}

	_, err = jsonparser.ArrayEach(body, func(value []byte, _ jsonparser.ValueType, _ int, _ error) {
		jobID, err := jsonparser.GetString(value, "id")
		if err != nil {
			slog.ErrorContext(ctx, "Error parsing job ID from jobs array", slog.Any("error", err))
			return
		}

		job, jerr := ScrapeJob(ctx, companyName, jobID)
		if jerr != nil {
			slog.ErrorContext(ctx, "Error parsing BambooHR job from jobs array", slog.Any("error", jerr))
			return
		}

		slog.DebugContext(ctx, "Parsed job", slog.String("job_id", job.SourceID), slog.String("title", job.Title))
		jobs = append(jobs, job)
	}, "result")
	if err != nil {
		slog.ErrorContext(ctx, "Error parsing jobs array from BambooHR job board endpoint", slog.Any("error", err))
		return jobs, fmt.Errorf("error parsing jobs array: %w", err)
	}

	return jobs, nil
}

// ScrapeJob scrapes an individual job from BambooHR ATS given the company name and job ID.
func ScrapeJob(ctx context.Context, companyName, jobID string) (*models.Job, error) {
	slog.DebugContext(ctx, "Scraping individual job", slog.String("ats", "bamboo"), slog.String("company_name", companyName), slog.String("job_id", jobID))

	jobURL := fmt.Sprintf("https://%s.bamboohr.com/careers/%s/detail", companyName, jobID)

	body, err := helpers.GetJSON(ctx, jobURL, nil)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting JSON from BambooHR job endpoint", slog.String("url", jobURL), slog.Any("error", err))
		return nil, fmt.Errorf("error getting JSON from BambooHR job endpoint: %w", err)
	}

	return parseBambooJob(ctx, body)
}

func parseBambooJob(ctx context.Context, data []byte) (*models.Job, error) {
	job := &models.Job{
		Source:     "bamboo",
		Department: models.Unsure,
	}
	job.SetSourceData(data)

	err := jsonparser.ObjectEach(data, func(key []byte, value []byte, _ jsonparser.ValueType, _ int) error {
		switch string(key) {
		case "jobOpeningShareUrl":
			// weirdly, this is over escaped like https:\/\/beehiiv.bamboohr.com\/careers\/25
			// so we need to unescape it
			url := strings.ReplaceAll(string(value), `\/`, `/`)
			job.URL = url
			// extract job ID from URL
			parts := strings.Split(url, "/")
			if len(parts) > 0 {
				job.SourceID = parts[len(parts)-1]
			}
		case "jobOpeningName":
			job.Title = string(value)
		case "departmentLabel":
			job.Department = models.ParseDepartment(string(value))
			job.DepartmentRaw = string(value)
		case "employmentStatusLabel":
			job.EmploymentType = models.ParseEmploymentType(string(value))
		case "description":
			job.Description = string(value)
		case "datePosted":
			job.ProcessDatePosted(ctx, value)
		case "compensation":
			compensation := models.ParseCompensation(string(value))
			if compensation.Parsed {
				job.CompensationUnit = compensation.Currency
				job.MinCompensation = compensation.MinSalary
				job.MaxCompensation = compensation.MaxSalary
			}

			job.AddMetadata("compensation", string(value))
		case "locationType":
			// "0" = in-office, "1" = remote, "2" = hybrid
			switch string(value) {
			case "0":
				job.LocationType = models.OnsiteLocation
			case "1":
				job.LocationType = models.RemoteLocation
			case "2":
				job.LocationType = models.HybridLocation
			default:
				job.LocationType = models.UnknownLocation
			}
		case "location":
			location := models.ParseLocation(value)

			if job.Location == "" {
				job.Location = location.String()
			}

			job.AddMetadata("location", location.String())
		case "atsLocation":
			location := models.ParseLocation(value)

			if job.Location == "" {
				job.Location = location.String()
			}

			job.AddMetadata("atsLocation", location.String())
		default:
			job.AddMetadata(string(key), string(value))
		}

		return nil
	}, "result", "jobOpening")
	if err != nil {
		slog.ErrorContext(ctx, "Error parsing job from BambooHR job endpoint", slog.Any("error", err))
		return nil, fmt.Errorf("error parsing job from BambooHR job endpoint: %w", err)
	}

	return job, nil
}
