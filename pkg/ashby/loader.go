package ashby

import (
	"context"
	"fmt"
	"time"

	"github.com/amalgamated-tools/jobscraping/pkg/helpers"
	"github.com/amalgamated-tools/jobscraping/pkg/models"
	"github.com/buger/jsonparser"
)

var ashbyCompanyURL = "https://api.ashbyhq.com/posting-api/job-board/%s?includeCompensation=true"

func ScrapeCompany(ctx context.Context, companyName string) error {
	companyURL := fmt.Sprintf(ashbyCompanyURL, companyName)
	body, err := helpers.GetJSON(companyURL)
	if err != nil {
		return fmt.Errorf("error getting JSON from Ashby job board endpoint: %w", err)
	}

	fmt.Printf("Received response: %s\n", string(body))

	_, err = jsonparser.ArrayEach(body, func(value []byte, dataType jsonparser.ValueType, offset int, ierr error) {
		job, jerr := parseAshbyJob(value)
		if jerr != nil {
			fmt.Printf("Error parsing job: %v\n", jerr)
			return
		}
		fmt.Printf("Parsed job: %+v\n", job)
	}, "jobs")

	if err != nil {
		return fmt.Errorf("error parsing jobs array: %w", err)
	}
	return nil
}

func parseAshbyJob(data []byte) (*models.Job, error) {
	job := &models.Job{
		Source: "ashby",
	}
	job.SetSourceData(data)

	err := jsonparser.ObjectEach(data, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		switch string(key) {
		case "id":
			job.SourceID = string(value)
		case "title":
			job.Title = string(value)
		case "department":
			job.Department = models.ParseDepartment(string(value))
		case "team":
			job.AddMetadata("team", string(value))
		case "employmentType":
			job.EmploymentType = models.ParseEmploymentType(string(value))
		case "location":
			job.Location = string(value)
		case "secondaryLocations":
			// this is an array of location objects
			_, jerr := jsonparser.ArrayEach(value, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				location, ierr := jsonparser.GetString(value, "location")
				if ierr != nil {
					fmt.Printf("error parsing secondary location: %v\n", err)
					return
				}
				job.AddMetadata("secondary_location", location)
			})
			if jerr != nil {
				return fmt.Errorf("error parsing secondaryLocations: %w", jerr)
			}
		case "publishedAt":
			stringValue, err := jsonparser.ParseString(value)
			if err != nil {
				return fmt.Errorf("error parsing publishedAt: %w", err)
			}
			datePosted, err := time.Parse("2006-01-02", stringValue)
			if err != nil {
				// if this is a time.ParseError, we can try to parse it as a full date-time
				datePosted, err = time.Parse(time.RFC3339, stringValue)
				if err == nil {
					job.DatePosted = datePosted.In(time.UTC) // Ensure the date is in UTC
				} else {
					return fmt.Errorf("error parsing publishedAt as date-time: %w", err)
				}
			} else {
				job.DatePosted = datePosted.In(time.UTC) // Ensure the date is in UTC
			}
			// publishedAt is a timestamp string "2025-11-14T00:33:59.437+00:00"
			fmt.Println("publishedAt")
		case "isRemote":
			isRemote, err := jsonparser.ParseBoolean(value)
			if err != nil {
				return fmt.Errorf("error parsing isRemote: %w", err)
			}
			job.IsRemote = isRemote
		case "jobUrl":
			job.URL = string(value)
		case "descriptionHtml":
			job.Description = string(value)
		case "descriptionPlain":
			// we can ignore this, prefer HTML version
			job.AddMetadata("description_plain", string(value))
		case "compensation":
			summary, err := jsonparser.GetString(value, "compensationTierSummary")
			if err != nil {
				return fmt.Errorf("error parsing compensation summary: %w", err)
			}
			// summary is like $155K - $190K or €185K - €317K
			comp := helpers.ParseCompensation(summary)
			if !comp.Parsed {
				return fmt.Errorf("unable to parse compensation string: %s", summary)
			}
			if comp.MinSalary != nil {
				job.MinCompensation = float64(*comp.MinSalary)
			}
			if comp.MaxSalary != nil {
				job.MaxCompensation = float64(*comp.MaxSalary)
			}
			if comp.Currency != "" {
				job.CompensationUnit = helpers.Ptr(comp.Currency)
			}
			if comp.OffersEquity {
				job.Equity = models.EquityOffered
			}
		default:
			job.AddMetadata(string(key), string(value))
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error parsing job object: %w", err)
	}

	return job, nil
}
