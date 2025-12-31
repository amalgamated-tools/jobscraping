// Package ashby implements an ATS loader for Ashby ATS.
package ashby

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
	ashbyCompanyURL = "https://api.ashbyhq.com/posting-api/job-board/%s?includeCompensation=true"
	ashbyJobQuery   = `{"query":"{\n\tjobPosting(organizationHostedJobsPageName: \"%s\", jobPostingId: \"%s\") {\ncompensationPhilosophyHtml\ncompensationTiers {\n  id\n  title\n  tierSummary\n}\ncompensationTierSummary\ndepartmentName\ndescriptionHtml\nemploymentType\nid\nisConfidential\nisListed\nlinkedData\nlocationAddress\nlocationName\npublishedDate\nscrapeableCompensationSalarySummary\nsecondaryLocationNames\nteamNames\ntitle\nworkplaceType\n\t}\n}"}`
)

// ScrapeCompany scrapes all jobs for a given company from Ashby ATS.
func ScrapeCompany(ctx context.Context, companyName string) ([]*models.Job, error) {
	slog.DebugContext(ctx, "Scraping company", slog.String("ats", "ashby"), slog.String("company_name", companyName))

	jobs := make([]*models.Job, 0)

	// The URL is like https://api.ashbyhq.com/posting-api/job-board/{companyName}?includeCompensation=true
	companyURL := fmt.Sprintf(ashbyCompanyURL, companyName)

	// Get the JSON from the company job board endpoint
	body, err := helpers.GetJSON(ctx, companyURL, nil)
	if err != nil {
		slog.ErrorContext(ctx, "Error getting JSON from Ashby job board endpoint", slog.String("url", companyURL), slog.Any("error", err))
		return jobs, fmt.Errorf("error getting JSON from Ashby job board endpoint: %w", err)
	}

	_, err = jsonparser.ArrayEach(body, func(value []byte, _ jsonparser.ValueType, _ int, _ error) {
		id, err := jsonparser.GetString(value, "id")
		if err != nil {
			slog.ErrorContext(ctx, "Error parsing job id from Ashby job board endpoint", slog.Any("error", err))
			return
		}

		job, err := ScrapeJob(ctx, companyName, id)
		if err != nil {
			slog.ErrorContext(ctx, "Error scraping individual job", slog.String("job_id", job.SourceID), slog.Any("error", err))
			return
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

// ScrapeJob scrapes an individual job from Ashby ATS given the company name and job ID.
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
		return nil, fmt.Errorf("error posting JSON to Ashby job endpoint: %w", err)
	}

	job, err := parseAshbyJob(ctx, bodyText)
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

	err := jsonparser.ObjectEach(job.GetSourceData(), func(key []byte, value []byte, _ jsonparser.ValueType, _ int) error {
		switch string(key) {
		case "linkedData":
			// linkedData is a JSON-LD object, we can try to parse it for more info
			jerr := jsonparser.ObjectEach(value, func(ldKey []byte, ldValue []byte, _ jsonparser.ValueType, _ int) error {
				switch string(ldKey) {
				case "hiringOrganization":
					orgName, err := jsonparser.GetString(ldValue, "name")
					if err == nil {
						job.Company.Name = orgName
					}

					sameAs, err := jsonparser.GetString(ldValue, "sameAs")
					if err == nil {
						job.Company.HomepageURL = helpers.Ptr(sameAs)
					}

					logo, err := jsonparser.GetString(ldValue, "logo")
					if err == nil {
						job.Company.LogoURL = helpers.Ptr(logo)
					}
				default:
					job.AddMetadata("linked_data_"+string(ldKey), string(ldValue))
				}

				return nil
			})
			if jerr != nil {
				slog.ErrorContext(ctx, "Error parsing linkedData object", slog.Any("error", jerr))
				return fmt.Errorf("error parsing linkedData object: %w", jerr)
			}
		case "id":
			job.SourceID = string(value)
		case "title":
			job.Title = string(value)
		case "locationName":
			job.Location = string(value)
		case "departmentName":
			job.Department = models.ParseDepartment(string(value))
			job.DepartmentRaw = string(value)
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
			_, jerr := jsonparser.ArrayEach(value, func(value []byte, _ jsonparser.ValueType, _ int, _ error) {
				job.AddMetadata("secondary_location", string(value))
			})
			if jerr != nil {
				slog.ErrorContext(ctx, "Error parsing secondaryLocationNames", slog.Any("error", jerr))
				return fmt.Errorf("error parsing secondaryLocationNames: %w", jerr)
			}
		case "publishedDate":
			job.ProcessDatePosted(ctx, value)
		case "teamNames":
			_, jerr := jsonparser.ArrayEach(value, func(value []byte, _ jsonparser.ValueType, _ int, _ error) {
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
				return nil // this is okay
			}

			job.MinCompensation = comp.MinSalary
			job.MaxCompensation = comp.MaxSalary

			if comp.Currency != "" {
				job.CompensationUnit = comp.Currency
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
		return job, fmt.Errorf("error parsing jobPosting object: %w", err)
	}

	return job, nil
}
