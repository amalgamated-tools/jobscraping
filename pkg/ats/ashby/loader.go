package ashby

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/amalgamated-tools/jobscraping/pkg/helpers"
	"github.com/amalgamated-tools/jobscraping/pkg/models"
	"github.com/buger/jsonparser"
)

var ashbyCompanyURL = "https://api.ashbyhq.com/posting-api/job-board/%s?includeCompensation=true"

var ashbyJobQuery = `{"query":"{\n\tjobPosting(organizationHostedJobsPageName: \"%s\", jobPostingId: \"%s\") {\ncompensationPhilosophyHtml\ncompensationTiers {\n  id\n  title\n  tierSummary\n}\ncompensationTierSummary\ndepartmentName\ndescriptionHtml\nemploymentType\nid\nisConfidential\nisListed\nlinkedData\nlocationAddress\nlocationName\npublishedDate\nscrapeableCompensationSalarySummary\nsecondaryLocationNames\nteamNames\ntitle\nworkplaceType\n\t}\n}"}`

func ScrapeCompany(ctx context.Context, companyName string, individual bool) error {
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

func ScrapeJob(ctx context.Context, companyName, jobID string) error {
	payload := strings.NewReader(
		fmt.Sprintf(ashbyJobQuery, companyName, jobID),
	)
	bodyText, err := helpers.PostJSON(
		"https://jobs.ashbyhq.com/api/non-user-graphql?op=ApiJobPosting",
		payload,
	)
	if err != nil {
		return err
	}

	job, err := parseAshbyJob([]byte(bodyText))
	if err != nil {
		return fmt.Errorf("error parsing Ashby job: %w", err)
	}

	job.URL = fmt.Sprintf("https://jobs.ashbyhq.com/%s/%s", companyName, jobID)
	return nil
}

func parseAshbyJob(data []byte) (*models.Job, error) {
	job := &models.Job{
		Source: "ashby",
	}
	job.SetSourceData(data)

	// see if this has data->jobPosting or not
	_, _, _, err := jsonparser.Get(data, "data", "jobPosting")
	if err == nil {
		// has data->jobPosting
		err = parseSingleJob(job)
		if err != nil {
			return nil, fmt.Errorf("error parsing Ashby jobPosting object: %w", err)
		}
		return job, nil
	} else {
		// no data->jobPosting, so parse directly
		err = parseCompanyJob(job)
		if err != nil {
			return nil, fmt.Errorf("error parsing Ashby job object: %w", err)
		}
		return job, nil
	}
}

func parseCompanyJob(job *models.Job) error {
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
		return fmt.Errorf("error parsing job object: %w", err)
	}

	return nil
}

func parseSingleJob(job *models.Job) error {
	err := jsonparser.ObjectEach(job.GetSourceData(), func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		fmt.Println(string(key))
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
				return fmt.Errorf("error parsing secondaryLocationNames: %w", jerr)
			}
		case "publishedDate":
			stringValue, err := jsonparser.ParseString(value)
			if err != nil {
				return fmt.Errorf("error parsing publishedDate: %w", err)
			}
			datePosted, err := time.Parse("2006-01-02", stringValue)
			if err != nil {
				// if this is a time.ParseError, we can try to parse it as a full date-time
				datePosted, err = time.Parse(time.RFC3339, stringValue)
				if err == nil {
					job.DatePosted = datePosted.In(time.UTC) // Ensure the date is in UTC
				} else {
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
				return fmt.Errorf("error parsing teamNames: %w", jerr)
			}
		case "compensationTierSummary":
			// summary is like $155K - $190K or €185K - €317K
			comp := helpers.ParseCompensation(string(value))
			if !comp.Parsed {
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
		return fmt.Errorf("error parsing jobPosting object: %w", err)
	}
	return nil
}
