// Package gem implements a loader for the Gem ATS.
package gem

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/amalgamated-tools/jobscraping/pkg/ats/models"
	"github.com/amalgamated-tools/jobscraping/pkg/helpers"
	"github.com/buger/jsonparser"
)

var (
	gemJobQuery = "[\n  {\n    \"operationName\": \"ExternalJobPostingQuery\",\n    \"variables\": {\n      \"boardId\": \"%s\",\n      \"extId\": \"%s\"\n    },\n    \"query\": \"fragment ExternalJobPostFragment on PublicOatsJobPost {\\n  id\\n  title\\n  descriptionHtml\\n  extId\\n  startDateTs\\n  firstPublishedTsSec\\n  companyLogo\\n  companyUrl\\n  isApplicationFormHidden\\n  applicationFormTemplate {\\n    id\\n    includeEeoc\\n    eeocConfig {\\n      includeRaceXGender\\n      includeVeteranStatus\\n      includeDisabilityStatus\\n      __typename\\n    }\\n    __typename\\n  }\\n  isUnlistedExternally\\n  locations {\\n    id\\n    name\\n    city\\n    isoCountry\\n    isRemote\\n    extId\\n    __typename\\n  }\\n  job {\\n    id\\n    locationType\\n    employmentType\\n    requisitionId\\n    teamDisplayName\\n    department {\\n      id\\n      name\\n      extId\\n      __typename\\n    }\\n    locations {\\n      id\\n      name\\n      city\\n      isoCountry\\n      isRemote\\n      extId\\n      __typename\\n    }\\n    __typename\\n  }\\n  jobPostSectionHtml {\\n    introHtml\\n    outroHtml\\n    __typename\\n  }\\n  __typename\\n}\\n\\nquery ExternalJobPostingQuery($boardId: String!, $extId: String!) {\\n  oatsExternalJobPosting(boardId: $boardId, extId: $extId) {\\n    id\\n    ...ExternalJobPostFragment\\n    __typename\\n  }\\n  oatsJobPostFieldsAndQuestions(\\n    jobBoardVanityPath: $boardId\\n    jobPostExtId: $extId\\n  ) {\\n    fields {\\n      fieldType\\n      isRequired\\n      __typename\\n    }\\n    questions {\\n      extId\\n      answerType\\n      displayType\\n      fileType\\n      text\\n      description\\n      isRequired\\n      options {\\n        extId\\n        value\\n        __typename\\n      }\\n      __typename\\n    }\\n    __typename\\n  }\\n}\\n\"\n  }\n]\n"
	gemGql      = "https://jobs.gem.com/api/public/graphql/batch"
	gemURL      = "https://api.gem.com/job_board/v0/%s/job_posts/"
)

// ScrapeCompany scrapes all job postings for a given company from the Gem ATS.
func ScrapeCompany(ctx context.Context, companyName string) ([]*models.Job, error) {
	slog.DebugContext(ctx, "Scraping company", slog.String("ats", "gem"), slog.String("company_name", companyName))

	jobs := make([]*models.Job, 0)

	body, err := helpers.GetJSON(
		ctx,
		fmt.Sprintf(gemURL, companyName),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting JSON from Gem job board endpoint: %w", err)
	}

	_, err = jsonparser.ArrayEach(body, func(value []byte, _ jsonparser.ValueType, _ int, _ error) {
		job, jerr := parseGemCompanyJob(ctx, value)
		if jerr != nil {
			slog.ErrorContext(ctx, "Error parsing Gem job from jobs array", slog.Any("error", jerr))
			return
		}

		slog.DebugContext(ctx, "Parsed job", slog.String("job_id", job.SourceID), slog.String("title", job.Title))
		jobs = append(jobs, job)
	})
	if err != nil {
		slog.ErrorContext(ctx, "Error parsing jobs array from Gem job board endpoint", slog.Any("error", err))
		return jobs, fmt.Errorf("error parsing jobs array: %w", err)
	}

	return jobs, nil
}

// ScrapeJob scrapes a specific job posting by job ID for a given company from the Gem ATS.
func ScrapeJob(ctx context.Context, companyName, jobID string) (*models.Job, error) {
	payload := strings.NewReader(
		fmt.Sprintf(gemJobQuery, companyName, jobID),
	)

	bodyText, err := helpers.PostJSON(
		ctx,
		gemGql,
		payload,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting JSON from Gem job board endpoint: %w", err)
	}

	job, err := parseGemOatsJob(ctx, bodyText)
	if err != nil {
		return nil, fmt.Errorf("error parsing Gem job object: %w", err)
	}

	job.URL = fmt.Sprintf("https://jobs.gem.com/%s/%s", companyName, jobID)

	return job, nil
}

func parseGemCompanyJob(ctx context.Context, data []byte) (*models.Job, error) {
	job := models.NewJob("gem", data)

	err := jsonparser.ObjectEach(job.GetSourceData(), func(key []byte, value []byte, _ jsonparser.ValueType, _ int) error {
		switch string(key) {
		case "absolute_url":
			job.URL = string(value)
		case "content":
			job.Description = string(value)
		case "created_at":
			job.AddMetadata("created_at", string(value))
		case "departments":
			_, _ = jsonparser.ArrayEach(value, func(deptValue []byte, _ jsonparser.ValueType, _ int, _ error) {
				// get the name
				deptName, err := jsonparser.GetString(deptValue, "name")
				if err != nil {
					slog.ErrorContext(ctx, "Error getting department name", slog.Any("error", err))
					return
				}

				if job.DepartmentRaw == "" || job.Department == models.UnknownDepartment {
					// we haven't set it yet
					job.DepartmentRaw = deptName
					job.Department = models.ParseDepartment(deptName)
				}

				job.AddMetadata("department", deptName)
			})
		case "employment_type":
			employmentType := string(value)
			job.EmploymentType = models.ParseEmploymentType(employmentType)
		case "first_published_at":
			job.ProcessDatePosted(ctx, value)
		case "id":
			job.AddMetadata("gem_id", string(value))
		case "internal_job_id":
			job.SourceID = string(value)
		case "location":
			locationName, err := jsonparser.GetString(value, "name")
			if err == nil {
				job.Location = locationName
			}
		case "location_type":
			job.LocationType = models.ParseLocationType(string(value))
		case "offices":
			_, _ = jsonparser.ArrayEach(value, func(locValue []byte, _ jsonparser.ValueType, _ int, _ error) {
				// get the name
				locName, err := jsonparser.GetString(locValue, "name")
				if err == nil {
					job.AddMetadata("office_location", locName)
				}

				locationName, err := jsonparser.GetString(locValue, "location", "name")
				if err == nil {
					job.AddMetadata("office_location_name", locationName)
				}
			})
		case "requisition_id":
			job.AddMetadata("requisition_id", string(value))
		case "title":
			job.Title = string(value)
		case "updated_at":
			job.AddMetadata("updated_at", string(value))
		default:
			job.AddMetadata(string(key), string(value))
		}

		return nil
	})
	if err != nil {
		slog.ErrorContext(ctx, "Error parsing Gem job object", slog.Any("error", err))
		return nil, fmt.Errorf("error parsing Gem job object: %w", err)
	}

	return job, nil
}

func parseGemOatsJob(ctx context.Context, data []byte) (*models.Job, error) {
	job := models.NewJob("gem", data)

	err := jsonparser.ObjectEach(job.GetSourceData(), func(key []byte, value []byte, _ jsonparser.ValueType, _ int) error {
		switch string(key) {
		case "id":
			job.AddMetadata("gem_id", string(value))
		case "title":
			job.Title = string(value)
		case "descriptionHtml":
			job.Description = string(value)
		case "extId":
			job.SourceID = string(value)
		case "firstPublishedTsSec":
			job.ProcessDatePosted(ctx, value)
		case "locations":
			_, _ = jsonparser.ArrayEach(value, func(locValue []byte, _ jsonparser.ValueType, _ int, _ error) {
				// get the name
				locName, err := jsonparser.GetString(locValue, "name")
				if err == nil && job.Location == "" {
					job.Location = locName
				}

				cityName, err := jsonparser.GetString(locValue, "city")
				if err == nil {
					job.AddMetadata("location_city", cityName)
				}

				countryCode, err := jsonparser.GetString(locValue, "isoCountry")
				if err == nil {
					job.AddMetadata("location_country", countryCode)
				}

				isRemote, err := jsonparser.GetBoolean(locValue, "isRemote")
				if err == nil {
					job.AddMetadata("location_is_remote", strconv.FormatBool(isRemote))
				}
			})
		case "job":
			locationType, err := jsonparser.GetString(value, "locationType")
			if err == nil {
				job.LocationType = models.ParseLocationType(locationType)
			}

			employmentType, err := jsonparser.GetString(value, "employmentType")
			if err == nil {
				job.EmploymentType = models.ParseEmploymentType(employmentType)
			}

			deptName, err := jsonparser.GetString(value, "department", "name")
			if err == nil {
				job.DepartmentRaw = deptName
				job.Department = models.ParseDepartment(deptName)
			}

			teamDisplayName, err := jsonparser.GetString(value, "teamDisplayName")
			if err == nil {
				job.AddMetadata("team_display_name", teamDisplayName)
			}

			_, _ = jsonparser.ArrayEach(value, func(locValue []byte, _ jsonparser.ValueType, _ int, _ error) {
				// get the name
				locName, err := jsonparser.GetString(locValue, "name")
				if err == nil {
					job.AddMetadata("job_location_name", locName)
				}

				cityName, err := jsonparser.GetString(locValue, "city")
				if err == nil {
					job.AddMetadata("job_location_city", cityName)
				}

				countryCode, err := jsonparser.GetString(locValue, "isoCountry")
				if err == nil {
					job.AddMetadata("job_location_country", countryCode)
				}

				isRemote, err := jsonparser.GetBoolean(locValue, "isRemote")
				if err == nil {
					job.AddMetadata("job_location_is_remote", strconv.FormatBool(isRemote))
				}
			})
		default:
			job.AddMetadata(string(key), string(value))
		}

		return nil
	})
	if err != nil {
		slog.ErrorContext(ctx, "Error parsing Gem job object", slog.Any("error", err))
		return nil, fmt.Errorf("error parsing Gem job object: %w", err)
	}

	return job, nil
}
