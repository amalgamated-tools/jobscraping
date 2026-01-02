// Package ashby implements an ATS loader for Ashby ATS.
package ashby

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

var (
	ashbyCompanyURL = "https://api.ashbyhq.com/posting-api/job-board/%s?includeCompensation=true"

	ashbyJobURL   = "https://jobs.ashbyhq.com/api/non-user-graphql?op=ApiJobPosting"
	ashbyJobQuery = `{"query":"{\n\tjobPosting(organizationHostedJobsPageName: \"%s\", jobPostingId: \"%s\") {\ncompensationPhilosophyHtml\ncompensationTiers {\n  id\n  title\n  tierSummary\n}\ncompensationTierSummary\ndepartmentName\ndescriptionHtml\nemploymentType\nid\nisConfidential\nisListed\nlinkedData\nlocationAddress\nlocationName\npublishedDate\nscrapeableCompensationSalarySummary\nsecondaryLocationNames\nteamNames\ntitle\nworkplaceType\n\t}\n}"}`

	ashbyCompanyInfoURL   = "https://jobs.ashbyhq.com/api/non-user-graphql?op=ApiOrganizationFromHostedJobsPageName"
	ashbyCompanyInfoQuery = "{\"query\":\"query ApiOrganizationFromHostedJobsPageName {\\n  organization: organizationFromHostedJobsPageName(\\n    organizationHostedJobsPageName: \\\"%s\\\"\\n    searchContext: JobBoard\\n  ) {\\n    ...OrganizationParts\\n    __typename\\n  }\\n}\\n\\nfragment OrganizationParts on Organization {\\n  name\\n  publicWebsite\\n  timezone\\n  theme {\\n    logoSquareImageUrl\\n  }\\n  __typename\\n}\"}"
)

// ScrapeCompany scrapes all jobs for a given company from Ashby ATS.
func ScrapeCompany(ctx context.Context, companyName string) ([]*models.Job, error) {
	slog.DebugContext(ctx, "Scraping company", slog.String("ats", "ashby"), slog.String("company_name", companyName))

	company, err := ScrapeCompanyInfo(ctx, companyName)
	if err != nil {
		slog.ErrorContext(ctx, "Error scraping company info for Ashby company", slog.String("company_name", companyName), slog.Any("error", err))
		return nil, fmt.Errorf("error scraping company info for Ashby company: %w", err)
	}

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

		if job.Company.Name == "" {
			job.Company = company
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
		ashbyJobURL,
		payload,
		nil,
	)
	if err != nil {
		slog.ErrorContext(ctx, "Error posting JSON to Ashby job endpoint", slog.String("url", ashbyJobURL), slog.Any("error", err))
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

// ScrapeCompanyInfo scrapes company information and jobs for a given company from Ashby ATS.
func ScrapeCompanyInfo(ctx context.Context, companyName string) (*models.Company, error) {
	slog.DebugContext(ctx, "Scraping company info", slog.String("ats", "ashby"), slog.String("company_name", companyName))

	payload := strings.NewReader(
		fmt.Sprintf(ashbyCompanyInfoQuery, companyName),
	)

	bodyText, err := helpers.PostJSON(
		ctx,
		ashbyCompanyInfoURL,
		payload,
		nil,
	)
	if err != nil {
		slog.ErrorContext(ctx, "Error posting JSON to Ashby company info endpoint", slog.String("url", ashbyCompanyInfoURL), slog.Any("error", err))
		return nil, fmt.Errorf("error posting JSON to Ashby company info endpoint: %w", err)
	}

	company := models.NewCompany()

	err = jsonparser.ObjectEach(bodyText, func(key []byte, value []byte, _ jsonparser.ValueType, _ int) error {
		switch string(key) {
		case "name":
			company.Name = string(value)
		case "publicWebsite":
			if string(value) != "" && !strings.EqualFold(string(value), "null") {
				homepage, err := url.Parse(string(value))
				if err == nil {
					company.Homepage = *homepage
				}
			}
		case "theme":
			logoURL, err := jsonparser.GetString(value, "logoSquareImageUrl")
			if err == nil && logoURL != "" && !strings.EqualFold(logoURL, "null") {
				logo, err := url.Parse(logoURL)
				if err == nil {
					company.Logo = *logo
				}
			}
		}

		return nil
	}, "data", "organization")
	if err != nil {
		slog.ErrorContext(ctx, "Error parsing organization object from Ashby company info endpoint", slog.Any("error", err))
		return company, fmt.Errorf("error parsing organization object: %w", err)
	}

	return company, nil
}

func parseAshbyJob(ctx context.Context, data []byte) (*models.Job, error) {
	job := models.NewJob("ashby", data)

	err := jsonparser.ObjectEach(job.GetSourceData(), func(key []byte, value []byte, _ jsonparser.ValueType, _ int) error {
		switch string(key) {
		case "compensationTierSummary":
			// summary is like $155K - $190K or €185K - €317K
			summary := string(value)
			if summary == "" || strings.EqualFold(summary, "null") {
				return nil
			}

			comp := models.ParseCompensation(summary)
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
		case "departmentName":
			job.Department = models.ParseDepartment(string(value))
			job.DepartmentRaw = string(value)
		case "descriptionHtml":
			job.Description = string(value)
		case "employmentType":
			job.EmploymentType = models.ParseEmploymentType(string(value))
		case "id":
			job.SourceID = string(value)
		case "linkedData":
			// linkedData is a JSON-LD object, we can try to parse it for more info
			jerr := jsonparser.ObjectEach(value, func(ldKey []byte, ldValue []byte, _ jsonparser.ValueType, _ int) error {
				switch string(ldKey) {
				case "title":
					title := string(ldValue)
					if job.Title == "" {
						slog.DebugContext(ctx, "Setting job title from linkedData", slog.String("title", title))
						job.Title = title
					}

				case "hiringOrganization":
					orgName, err := jsonparser.GetString(ldValue, "name")
					if err == nil {
						job.Company.Name = orgName
					}

					sameAs, err := jsonparser.GetString(ldValue, "sameAs")
					if err == nil {
						homepage, err := url.Parse(sameAs)
						if err == nil {
							job.Company.Homepage = *homepage
						} else {
							slog.ErrorContext(ctx, "Error parsing company homepage URL from Ashby linkedData", slog.String("url", sameAs), slog.Any("error", err))
						}
					}

					logo, err := jsonparser.GetString(ldValue, "logo")
					if err == nil {
						logoURL, err := url.Parse(logo)
						if err == nil {
							job.Company.Logo = *logoURL
						} else {
							slog.ErrorContext(ctx, "Error parsing company logo URL from Ashby linkedData", slog.String("url", logo), slog.Any("error", err))
						}
					}
				case "jobLocation":
					location := models.ParseLocation(value)
					if job.Location == "" {
						job.Location = location.String()
					}

					job.AddMetadata("linked_data_location", location.String())
				default:
					job.AddMetadata("linked_data_"+string(ldKey), string(ldValue))
				}

				return nil
			})
			if jerr != nil {
				slog.ErrorContext(ctx, "Error parsing linkedData object", slog.Any("error", jerr))
				return fmt.Errorf("error parsing linkedData object: %w", jerr)
			}
		case "locationName":
			job.Location = string(value)
		case "publishedDate":
			job.ProcessDatePosted(ctx, value)
		case "secondaryLocationNames":
			_, jerr := jsonparser.ArrayEach(value, func(value []byte, _ jsonparser.ValueType, _ int, _ error) {
				job.AddMetadata("secondary_location", string(value))
			})
			if jerr != nil {
				slog.ErrorContext(ctx, "Error parsing secondaryLocationNames", slog.Any("error", jerr))
				return fmt.Errorf("error parsing secondaryLocationNames: %w", jerr)
			}
		case "teamNames":
			_, jerr := jsonparser.ArrayEach(value, func(value []byte, _ jsonparser.ValueType, _ int, _ error) {
				job.AddMetadata("team", string(value))
			})
			if jerr != nil {
				slog.ErrorContext(ctx, "Error parsing teamNames", slog.Any("error", jerr))
				return fmt.Errorf("error parsing teamNames: %w", jerr)
			}
		case "title":
			job.Title = string(value)
		case "workplaceType":
			// possible values: REMOTE, HYBRID, ONSITE
			workplaceType, err := jsonparser.ParseString(value)
			if err != nil {
				slog.ErrorContext(ctx, "Error parsing workplaceType", slog.Any("error", err))
				return fmt.Errorf("error parsing workplaceType: %w", err)
			}

			if strings.EqualFold(workplaceType, "Remote") {
				job.IsRemote = true
				job.LocationType = models.RemoteLocation
			} else {
				job.IsRemote = false
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
