// Package lever provides functions to scrape job postings from Lever ATS.
package lever

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
	leverCompanyURL = "https://api.lever.co/v0/postings/%s?mode=json"
	leverJobURL     = "https://api.lever.co/v0/postings/%s/%s?mode=json"
)

// ScrapeCompany scrapes all jobs for a given company from Lever ATS.
func ScrapeCompany(ctx context.Context, companyName string) ([]*models.Job, error) {
	slog.DebugContext(ctx, "Scraping company", slog.String("ats", "lever"), slog.String("company_name", companyName))

	jobs := make([]*models.Job, 0)

	// The URL is like https://api.lever.co/v0/postings/{companyName}?mode=json
	companyURL := fmt.Sprintf(leverCompanyURL, companyName)

	// Get the JSON from the company job board endpoint
	body, err := helpers.GetJSON(ctx, companyURL, nil)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting JSON from Lever job board endpoint", slog.String("url", companyURL), slog.Any("error", err))
		return jobs, fmt.Errorf("error getting JSON from Lever job board endpoint: %w", err)
	}

	_, err = jsonparser.ArrayEach(body, func(value []byte, _ jsonparser.ValueType, _ int, _ error) {
		job, jerr := parseLeverJob(ctx, value)
		if jerr != nil {
			slog.ErrorContext(ctx, "Error parsing Lever job from jobs array", slog.Any("error", jerr))
			return
		}

		if job.Company.Name == "" {
			err := scrapeCompanyInfo(ctx, job)
			if err != nil {
				slog.ErrorContext(ctx, "Error scraping company info from job URL", slog.String("url", job.URL), slog.Any("error", err))
				// we continue even if there's an error here
			}
		}

		jobs = append(jobs, job)
	})
	if err != nil {
		slog.ErrorContext(ctx, "Error parsing jobs array from Lever job board endpoint", slog.Any("error", err))
		return jobs, fmt.Errorf("error parsing jobs array: %w", err)
	}

	return jobs, nil
}

// ScrapeJob scrapes an individual job from Lever ATS given the company name and job ID.
func ScrapeJob(ctx context.Context, companyName, jobID string) (*models.Job, error) {
	slog.DebugContext(ctx, "Scraping individual job", slog.String("ats", "lever"), slog.String("company_name", companyName), slog.String("job_id", jobID))

	// The URL is like https://api.lever.co/v0/postings/{companyName}/{jobID}?mode=json
	jobURL := fmt.Sprintf(leverJobURL, companyName, jobID)

	// Get the JSON from the job endpoint
	body, err := helpers.GetJSON(ctx, jobURL, nil)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting JSON from Lever job endpoint", slog.String("url", jobURL), slog.Any("error", err))
		return nil, fmt.Errorf("error getting JSON from Lever job endpoint: %w", err)
	}

	job, err := parseLeverJob(ctx, body)
	if err != nil {
		slog.ErrorContext(ctx, "Error parsing Lever job from job endpoint", slog.Any("error", err))
		return nil, fmt.Errorf("error parsing Lever job from job endpoint: %w", err)
	}
	// Try to get company info from LD+JSON on the job page
	err = scrapeCompanyInfo(ctx, job)
	if err != nil {
		slog.ErrorContext(ctx, "Error scraping company info from job URL", slog.String("url", job.URL), slog.Any("error", err))
		// we continue even if there's an error here
	}

	slog.DebugContext(ctx, "Parsed job", slog.String("job_id", job.SourceID), slog.String("title", job.Title))

	return job, nil
}

func scrapeCompanyInfo(ctx context.Context, job *models.Job) error {
	// Try to get company name and logo from LD+JSON on the job page
	json, err := helpers.GetLDJSON(ctx, job.URL)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting LD+JSON from job URL", slog.String("url", job.URL), slog.Any("error", err))
		return fmt.Errorf("error getting LD+JSON from job URL: %w", err)
	}

	org, ok := json["hiringOrganization"].(map[string]any)
	if ok {
		name, ok := org["name"].(string)
		if ok {
			job.Company.Name = name
		}
	}

	logo, ok := org["logo"].(string)
	if ok {
		logoURL, err := url.Parse(logo)
		if err == nil {
			job.Company.Logo = *logoURL
		}
	}

	return nil
}

func parseLeverJob(ctx context.Context, data []byte) (*models.Job, error) {
	job := models.NewJob("lever", data)

	err := jsonparser.ObjectEach(job.GetSourceData(), func(key []byte, value []byte, _ jsonparser.ValueType, _ int) error {
		switch string(key) {
		case "categories":
			// Object with location, commitment, team, department, and allLocations.
			// Note: primary posting location is represented by location, and also appears in the allLocations array.
			commitment, err := jsonparser.GetString(value, "commitment")
			if err != nil {
				slog.ErrorContext(ctx, "Error parsing commitment from categories", slog.Any("error", err))
				// we continue even if there's an error here
			}

			job.EmploymentType = models.ParseEmploymentType(commitment)
			job.AddMetadata("commitment_raw", commitment)

			location, err := jsonparser.GetString(value, "location")
			if err != nil {
				slog.ErrorContext(ctx, "Error parsing location from categories", slog.Any("error", err))
				// we continue even if there's an error here
			}

			job.Location = location
			job.AddMetadata("location_raw", location)

			department, err := jsonparser.GetString(value, "department")
			if err != nil {
				slog.ErrorContext(ctx, "Error parsing department from categories", slog.Any("error", err))
				// we continue even if there's an error here
			}

			job.Department = models.ParseDepartment(department)
			job.DepartmentRaw = department

			team, err := jsonparser.GetString(value, "team")
			if err != nil {
				slog.ErrorContext(ctx, "Error parsing team from categories", slog.Any("error", err))
				// we continue even if there's an error here
			}

			job.AddMetadata("team", team)

			_, err = jsonparser.ArrayEach(value, func(locValue []byte, _ jsonparser.ValueType, _ int, _ error) {
				// add each location to metadata
				job.AddMetadata("secondary_location", string(locValue))
			}, "allLocations")
			if err != nil {
				slog.ErrorContext(ctx, "Error parsing allLocations from categories", slog.Any("error", err))
				// we continue even if there's an error here
			}
		case "createdAt":
			job.ProcessDatePosted(ctx, value)
		case "descriptionPlain":
			job.Description = string(value)
		case "id":
			job.SourceID = string(value)
		case "lists":
			// Extra lists (such as requirements, benefits, etc.) from the job posting. This is a list of {text:NAME, content:"unstyled HTML of list elements"}
			_, err := jsonparser.ArrayEach(value, func(listValue []byte, _ jsonparser.ValueType, _ int, _ error) {
				listName, err := jsonparser.GetString(listValue, "text")
				if err != nil {
					slog.ErrorContext(ctx, "Error parsing list name from lists", slog.Any("error", err))
					return
				}

				listContent, err := jsonparser.GetString(listValue, "content")
				if err != nil {
					slog.ErrorContext(ctx, "Error parsing list content from lists", slog.Any("error", err))
					return
				}

				job.AddMetadata("list_"+listName, listContent)
			})
			if err != nil {
				slog.ErrorContext(ctx, "Error parsing lists array", slog.Any("error", err))
				// we continue even if there's an error here
			}
		case "salaryRange":
			// Object with currency, interval, min, and max. This field is optional. In XML mode this field is parsed into a string.
			minimum, err := jsonparser.GetFloat(value, "min")
			if err != nil {
				slog.ErrorContext(ctx, "Error parsing min from salaryRange", slog.Any("error", err))
				// we continue even if there's an error here
			}

			job.MinCompensation = minimum

			maximum, err := jsonparser.GetFloat(value, "max")
			if err != nil {
				slog.ErrorContext(ctx, "Error parsing max from salaryRange", slog.Any("error", err))
				// we continue even if there's an error here
			}

			job.MaxCompensation = maximum

			currency, err := jsonparser.GetString(value, "currency")
			if err != nil {
				slog.ErrorContext(ctx, "Error parsing currency from salaryRange", slog.Any("error", err))
				// we continue even if there's an error here
			}

			job.CompensationUnit = currency

			interval, err := jsonparser.GetString(value, "interval")
			if err != nil {
				slog.ErrorContext(ctx, "Error parsing interval from salaryRange", slog.Any("error", err))
				// we continue even if there's an error here
			}

			job.AddMetadata("compensation_interval", interval)
		case "text":
			job.Title = string(value)
		case "country":
			// An ISO 3166-1 alpha-2 code for a country / territory (or null to indicate an unknown country). This is not filterable.
			job.AddMetadata("country", string(value))
		case "workplaceType":
			// Describes the primary workplace environment for a job posting. May be one of unspecified, on-site, remote, or hybrid. Not filterable
			workplaceType := string(value)

			job.LocationType = models.ParseLocationType(workplaceType)
			if job.LocationType == models.RemoteLocation {
				job.IsRemote = true
			}
		case "hostedUrl":
			job.URL = string(value)
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
