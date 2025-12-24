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
	// payload := strings.NewReader(
	// 	fmt.Sprintf(ashbyJobQuery, companyName, jobID),
	// )

	// bodyText, err := helpers.PostJSON(
	// 	ctx,
	// 	"https://jobs.ashbyhq.com/api/non-user-graphql?op=ApiJobPosting",
	// 	payload,
	// 	nil,
	// )
	// if err != nil {
	// 	slog.ErrorContext(ctx, "Error posting JSON to Ashby job endpoint", slog.String("url", "https://jobs.ashbyhq.com/api/non-user-graphql?op=ApiJobPosting"), slog.Any("error", err))
	// 	return nil, fmt.Errorf("error posting JSON to Ashby job endpoint: %w", err)
	// }

	// job, err := parseAshbyJob(ctx, bodyText)
	// if err != nil {
	// 	slog.ErrorContext(ctx, "Error parsing Ashby job from individual job endpoint", slog.String("job_id", jobID), slog.Any("error", err))
	// 	return nil, fmt.Errorf("error parsing Ashby job: %w", err)
	// }

	// job.URL = fmt.Sprintf("https://jobs.ashbyhq.com/%s/%s", companyName, jobID)

	return nil, nil
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

// func parseSingleJob(ctx context.Context, job *models.Job) error {
// 	slog.DebugContext(ctx, "Parsing single jobPosting object")

// 	err := jsonparser.ObjectEach(job.GetSourceData(), func(key []byte, value []byte, _ jsonparser.ValueType, _ int) error {
// 		switch string(key) {
// 		case "id":
// 			job.SourceID = string(value)
// 		case "title":
// 			job.Title = string(value)
// 		case "locationName":
// 			job.Location = string(value)
// 		case "departmentName":
// 			job.Department = models.ParseDepartment(string(value))
// 		case "workplaceType":
// 			// possible values: REMOTE, HYBRID, ONSITE
// 			workplaceType, err := jsonparser.ParseString(value)
// 			if err != nil {
// 				slog.ErrorContext(ctx, "Error parsing workplaceType", slog.Any("error", err))
// 				return fmt.Errorf("error parsing workplaceType: %w", err)
// 			}

// 			if strings.EqualFold(workplaceType, "Remote") {
// 				job.IsRemote = true
// 			} else {
// 				job.IsRemote = false
// 			}
// 		case "secondaryLocationNames":
// 			_, jerr := jsonparser.ArrayEach(value, func(value []byte, _ jsonparser.ValueType, _ int, _ error) {
// 				job.AddMetadata("secondary_location", string(value))
// 			})
// 			if jerr != nil {
// 				slog.ErrorContext(ctx, "Error parsing secondaryLocationNames", slog.Any("error", jerr))
// 				return fmt.Errorf("error parsing secondaryLocationNames: %w", jerr)
// 			}
// 		case "publishedDate":
// 			processDatePosted(ctx, job, value)
// 		case "teamNames":
// 			_, jerr := jsonparser.ArrayEach(value, func(value []byte, _ jsonparser.ValueType, _ int, _ error) {
// 				job.AddMetadata("team", string(value))
// 			})
// 			if jerr != nil {
// 				slog.ErrorContext(ctx, "Error parsing teamNames", slog.Any("error", jerr))
// 				return fmt.Errorf("error parsing teamNames: %w", jerr)
// 			}
// 		case "compensationTierSummary":
// 			// summary is like $155K - $190K or €185K - €317K
// 			comp := helpers.ParseCompensation(string(value))
// 			if !comp.Parsed {
// 				slog.ErrorContext(ctx, "Unable to parse compensation string", slog.String("summary", string(value)))
// 				return fmt.Errorf("%w: %s", ErrUnableToParseCompensation, string(value))
// 			}

// 			if comp.MinSalary != nil {
// 				job.MinCompensation = float64(*comp.MinSalary)
// 			}

// 			if comp.MaxSalary != nil {
// 				job.MaxCompensation = float64(*comp.MaxSalary)
// 			}

// 			if comp.Currency != "" {
// 				job.CompensationUnit = helpers.Ptr(comp.Currency)
// 			}

// 			if comp.OffersEquity {
// 				job.Equity = models.EquityOffered
// 			}
// 		default:
// 			job.AddMetadata(string(key), string(value))
// 		}

// 		return nil
// 	}, "data", "jobPosting")
// 	if err != nil {
// 		slog.ErrorContext(ctx, "Error parsing jobPosting object", slog.Any("error", err))
// 		return fmt.Errorf("error parsing jobPosting object: %w", err)
// 	}

// 	return nil
// }

// func processDatePosted(ctx context.Context, job *models.Job, value []byte) {
// 	stringValue, err := jsonparser.ParseString(value)
// 	if err != nil {
// 		slog.ErrorContext(ctx, "Error parsing publishedDate", slog.Any("error", err))
// 		return
// 	}

// 	datePosted, err := time.Parse("2006-01-02", stringValue)
// 	if err != nil {
// 		// if this is a time.ParseError, we can try to parse it as a full date-time
// 		datePosted, err = time.Parse(time.RFC3339, stringValue)
// 		if err == nil {
// 			job.DatePosted = datePosted.In(time.UTC) // Ensure the date is in UTC
// 		} else {
// 			slog.ErrorContext(ctx, "Error parsing publishedDate as date-time", slog.Any("error", err))
// 			return
// 		}
// 	} else {
// 		job.DatePosted = datePosted.In(time.UTC) // Ensure the date is in UTC
// 	}
// }
