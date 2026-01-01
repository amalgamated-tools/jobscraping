// Package bamboo contains functions to scrape jobs from BambooHR ATS.
package bamboo

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
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

		if job.Company.Name == "" {
			company, err := ScrapeCompanyInfo(ctx, companyName)
			if err != nil {
				slog.ErrorContext(ctx, "Error scraping company info for BambooHR job", slog.String("company_name", companyName), slog.Any("error", err))
			} else {
				job.Company = company
			}
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

	job, err := parseBambooJob(ctx, body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse job %s for company %s: %w", jobID, companyName, err)
	}

	company, err := ScrapeCompanyInfo(ctx, companyName)
	if err != nil {
		slog.ErrorContext(ctx, "Error scraping company info for BambooHR job", slog.String("company_name", companyName), slog.Any("error", err))
	} else {
		job.Company = company
	}

	return job, nil
}

// ScrapeCompanyInfo scrapes company info from BambooHR ATS given the company name.
func ScrapeCompanyInfo(ctx context.Context, companyName string) (*models.Company, error) {
	slog.DebugContext(ctx, "Scraping company info", slog.String("ats", "bamboo"), slog.String("company_name", companyName))

	companyInfoURL := fmt.Sprintf("https://%s.bamboohr.com/careers/company-info", companyName)

	bodyText, err := helpers.GetJSON(ctx, companyInfoURL, nil)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting JSON from BambooHR company info endpoint", slog.String("url", companyInfoURL), slog.Any("error", err))
		return nil, fmt.Errorf("error getting JSON from BambooHR company info endpoint: %w", err)
	}

	company := models.NewCompany()

	err = jsonparser.ObjectEach(bodyText, func(key []byte, value []byte, _ jsonparser.ValueType, _ int) error {
		switch string(key) {
		case "name":
			company.Name = string(value)
		case "careerShareUrl":
			if string(value) != "" && !strings.EqualFold(string(value), "null") {
				homepage, err := url.Parse(string(value))
				if err == nil {
					company.Homepage = *homepage
				} else {
					slog.ErrorContext(ctx, "Error parsing company homepage URL from BambooHR company info endpoint", slog.String("url", string(value)), slog.Any("error", err))
				}
			}
		case "logoUrl":
			if string(value) != "" && !strings.EqualFold(string(value), "null") {
				logo, err := url.Parse(string(value))
				if err == nil {
					company.Logo = *logo
				} else {
					slog.ErrorContext(ctx, "Error parsing company logo URL from BambooHR company info endpoint", slog.String("url", string(value)), slog.Any("error", err))
				}
			}
		}

		return nil
	}, "result")
	if err != nil {
		slog.ErrorContext(ctx, "Error parsing organization object from Ashby company info endpoint", slog.Any("error", err))
		return company, fmt.Errorf("error parsing organization object: %w", err)
	}

	return company, nil
}

func parseBambooJob(ctx context.Context, data []byte) (*models.Job, error) {
	job := models.NewJob("bamboo", data)

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
				job.LocationType = models.UnknownLocationType
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
