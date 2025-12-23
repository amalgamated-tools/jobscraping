package ashby

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/amalgamated-tools/jobscraping/pkg/ats/models"
	"github.com/amalgamated-tools/jobscraping/pkg/helpers"
	"github.com/buger/jsonparser"
)

var ashbyCompanyURL = "https://api.ashbyhq.com/posting-api/job-board/%s?includeCompensation=true"

var ashbyJobQuery = `{"query":"{\n\tjobPosting(organizationHostedJobsPageName: \"%s\", jobPostingId: \"%s\") {\ncompensationPhilosophyHtml\ncompensationTiers {\n  id\n  title\n  tierSummary\n}\ncompensationTierSummary\ndepartmentName\ndescriptionHtml\nemploymentType\nid\nisConfidential\nisListed\nlinkedData\nlocationAddress\nlocationName\npublishedDate\nscrapeableCompensationSalarySummary\nsecondaryLocationNames\nteamNames\ntitle\nworkplaceType\n\t}\n}"}`

func ScrapeCompany(ctx context.Context, companyName string, scrapeIndividual bool) ([]*models.Job, error) {
	slog.DebugContext(ctx, "Scraping company", slog.String("ats", "ashby"), slog.String("company_name", companyName), slog.Bool("scrape_individual", scrapeIndividual))

	jobs := make([]*models.Job, 0)

	// The URL is like https://api.ashbyhq.com/posting-api/job-board/{companyName}?includeCompensation=true
	companyURL := fmt.Sprintf(ashbyCompanyURL, companyName)

	// Get the JSON from the company job board endpoint
	body, err := helpers.GetJSON(ctx, companyURL, nil)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting JSON from Ashby job board endpoint", slog.String("url", companyURL), slog.Any("error", err))
		return jobs, fmt.Errorf("error getting JSON from Ashby job board endpoint: %w", err)
	}

	_, err = jsonparser.ArrayEach(body, func(value []byte, dataType jsonparser.ValueType, offset int, ierr error) {
		job, jerr := parseAshbyJob(ctx, value)
		if jerr != nil {
			slog.ErrorContext(ctx, "Error parsing Ashby job from jobs array", slog.Any("error", jerr))
			return
		}

		if scrapeIndividual {
			slog.DebugContext(ctx, "Scraping individual job", slog.String("job_id", job.SourceID))

			job, err := ScrapeJob(ctx, companyName, job.SourceID)
			if err != nil {
				slog.ErrorContext(ctx, "Error scraping individual job", slog.String("job_id", job.SourceID), slog.Any("error", err))
				return
			}
		}

		slog.DebugContext(ctx, "Parsed job", slog.String("job_id", job.SourceID), slog.String("title", job.Title))
		jobs = append(jobs, job)
	}, "jobs")
	if err != nil {
		slog.ErrorContext(ctx, "Error parsing jobs array from Ashby job board endpoint", slog.Any("error", err))
		return jobs, fmt.Errorf("error parsing jobs array: %w", err)
	}

	return jobs, nil
}

func ScrapeJob(ctx context.Context, companyName, jobID string) (*models.Job, error) {
	slog.DebugContext(ctx, "Scraping individual job", slog.String("ats", "ashby"), slog.String("company_name", companyName), slog.String("job_id", jobID))
	payload := strings.NewReader(
		fmt.Sprintf(ashbyJobQuery, companyName, jobID),
	)

	bodyText, err := helpers.PostJSON(
		ctx,
		"https://jobs.ashbyhq.com/api/non-user-graphql?op=ApiJobPosting",
		payload,
		nil,
	)
	if err != nil {
		slog.ErrorContext(ctx, "Error posting JSON to Ashby job endpoint", slog.String("url", "https://jobs.ashbyhq.com/api/non-user-graphql?op=ApiJobPosting"), slog.Any("error", err))
		return nil, err
	}

	job, err := parseAshbyJob(ctx, []byte(bodyText))
	if err != nil {
		slog.ErrorContext(ctx, "Error parsing Ashby job from individual job endpoint", slog.String("job_id", jobID), slog.Any("error", err))
		return nil, fmt.Errorf("error parsing Ashby job: %w", err)
	}

	job.URL = fmt.Sprintf("https://jobs.ashbyhq.com/%s/%s", companyName, jobID)

	return job, nil
}

func parseAshbyJob(ctx context.Context, data []byte) (*models.Job, error) {
	job := &models.Job{
		Source: "ashby",
	}
	job.SetSourceData(data)

	// see if this has data->jobPosting or not
	_, _, _, err := jsonparser.Get(data, "data", "jobPosting")
	if err == nil {
		// has data->jobPosting
		slog.DebugContext(ctx, "Parsing single jobPosting object")

		err = parseSingleJob(ctx, job)
		if err != nil {
			slog.ErrorContext(ctx, "Error parsing single jobPosting object", slog.Any("error", err))
			return nil, fmt.Errorf("error parsing Ashby jobPosting object: %w", err)
		}

		return job, nil
	} else {
		// no data->jobPosting, so parse directly
		slog.DebugContext(ctx, "Parsing company job object")

		err = parseCompanyJob(ctx, job)
		if err != nil {
			slog.ErrorContext(ctx, "Error parsing company job object", slog.Any("error", err))
			return nil, fmt.Errorf("error parsing Ashby job object: %w", err)
		}

		return job, nil
	}
}

func parseCompanyJob(ctx context.Context, job *models.Job) error {
	slog.DebugContext(ctx, "Parsing company job object")

	err := jsonparser.ObjectEach(job.GetSourceData(), func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
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
					slog.ErrorContext(ctx, "Error parsing secondaryLocations location", slog.Any("error", ierr))
					return
				}

				job.AddMetadata("secondary_location", location)
			})
			if jerr != nil {
				slog.ErrorContext(ctx, "Error parsing secondaryLocations", slog.Any("error", jerr))
				return fmt.Errorf("error parsing secondaryLocations: %w", jerr)
			}
		case "publishedAt":
			stringValue, err := jsonparser.ParseString(value)
			if err != nil {
				slog.ErrorContext(ctx, "Error parsing publishedAt", slog.Any("error", err))
				return fmt.Errorf("error parsing publishedAt: %w", err)
			}

			datePosted, err := time.Parse("2006-01-02", stringValue)
			if err != nil {
				// if this is a time.ParseError, we can try to parse it as a full date-time
				datePosted, err = time.Parse(time.RFC3339, stringValue)
				if err == nil {
					job.DatePosted = datePosted.In(time.UTC) // Ensure the date is in UTC
				} else {
					slog.ErrorContext(ctx, "Error parsing publishedAt as date-time", slog.Any("error", err))
					return fmt.Errorf("error parsing publishedAt as date-time: %w", err)
				}
			} else {
				job.DatePosted = datePosted.In(time.UTC) // Ensure the date is in UTC
			}
			// publishedAt is a timestamp string "2025-11-14T00:33:59.437+00:00"

		case "isRemote":
			isRemote, err := jsonparser.ParseBoolean(value)
			if err != nil {
				slog.ErrorContext(ctx, "Error parsing isRemote", slog.Any("error", err))
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
				slog.ErrorContext(ctx, "Error parsing compensation summary", slog.Any("error", err))
				return fmt.Errorf("error parsing compensation summary: %w", err)
			}
			// summary is like $155K - $190K or €185K - €317K
			comp := helpers.ParseCompensation(summary)
			if !comp.Parsed {
				slog.ErrorContext(ctx, "Unable to parse compensation string", slog.String("summary", summary))
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
		slog.ErrorContext(ctx, "Error parsing job object", slog.Any("error", err))
		return fmt.Errorf("error parsing job object: %w", err)
	}

	return nil
}

func parseSingleJob(ctx context.Context, job *models.Job) error {
	slog.DebugContext(ctx, "Parsing single jobPosting object")

	err := jsonparser.ObjectEach(job.GetSourceData(), func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		switch string(key) {
		case "id":
			job.SourceID = string(value)
		case "title":
			job.Title = string(value)
		case "locationName":
			job.Location = string(value)
		case "departmentName":
			job.Department = models.ParseDepartment(string(value))
		case "workplaceType":
			// possible values: REMOTE, HYBRID, ONSITE
			workplaceType, err := jsonparser.ParseString(value)
			if err != nil {
				slog.ErrorContext(ctx, "Error parsing workplaceType", slog.Any("error", err))
				return fmt.Errorf("error parsing workplaceType: %w", err)
			}

			if strings.EqualFold(workplaceType, "Remote") {
				job.IsRemote = true
			} else {
				job.IsRemote = false
			}
		case "secondaryLocationNames":
			_, jerr := jsonparser.ArrayEach(value, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				job.AddMetadata("secondary_location", string(value))
			})
			if jerr != nil {
				slog.ErrorContext(ctx, "Error parsing secondaryLocationNames", slog.Any("error", jerr))
				return fmt.Errorf("error parsing secondaryLocationNames: %w", jerr)
			}
		case "publishedDate":
			stringValue, err := jsonparser.ParseString(value)
			if err != nil {
				slog.ErrorContext(ctx, "Error parsing publishedDate", slog.Any("error", err))
				return fmt.Errorf("error parsing publishedDate: %w", err)
			}

			datePosted, err := time.Parse("2006-01-02", stringValue)
			if err != nil {
				// if this is a time.ParseError, we can try to parse it as a full date-time
				datePosted, err = time.Parse(time.RFC3339, stringValue)
				if err == nil {
					job.DatePosted = datePosted.In(time.UTC) // Ensure the date is in UTC
				} else {
					slog.ErrorContext(ctx, "Error parsing publishedDate as date-time", slog.Any("error", err))
					return fmt.Errorf("error parsing publishedDate as date-time: %w", err)
				}
			} else {
				job.DatePosted = datePosted.In(time.UTC) // Ensure the date is in UTC
			}
		case "teamNames":
			_, jerr := jsonparser.ArrayEach(value, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				job.AddMetadata("team", string(value))
			})
			if jerr != nil {
				slog.ErrorContext(ctx, "Error parsing teamNames", slog.Any("error", jerr))
				return fmt.Errorf("error parsing teamNames: %w", jerr)
			}
		case "compensationTierSummary":
			// summary is like $155K - $190K or €185K - €317K
			comp := helpers.ParseCompensation(string(value))
			if !comp.Parsed {
				slog.ErrorContext(ctx, "Unable to parse compensation string", slog.String("summary", string(value)))
				return fmt.Errorf("unable to parse compensation string: %s", string(value))
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
	}, "data", "jobPosting")
	if err != nil {
		slog.ErrorContext(ctx, "Error parsing jobPosting object", slog.Any("error", err))
		return fmt.Errorf("error parsing jobPosting object: %w", err)
	}

	return nil
}
