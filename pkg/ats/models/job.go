package models

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/buger/jsonparser"
)

const millisecondsToSeconds = 1000

// Job represents a job posting with various attributes.
type Job struct {
	URL              string         `json:"url"`
	Company          Company        `json:"company"`
	CompensationUnit string         `json:"compensation_unit"`
	DatePosted       time.Time      `json:"date_posted"`
	Department       Department     `json:"department"`
	DepartmentRaw    string         `json:"department_raw,omitempty"`
	Description      string         `json:"description"`
	EmploymentType   EmploymentType `json:"employment_type,omitempty"`
	Equity           EquityType     `json:"equity,omitempty"`
	IsRemote         bool           `json:"is_remote"`
	Location         string         `json:"location,omitempty"`
	LocationType     LocationType   `json:"location_type,omitempty"`
	MaxCompensation  float64        `json:"max_compensation"`
	MinCompensation  float64        `json:"min_compensation"`
	Source           string         `json:"source"`
	SourceID         string         `json:"source_id"`
	Title            string         `json:"title"`

	Tags map[string][]string `json:"tags,omitempty"`

	sourceData []byte `json:"-"`
	// Company *Company `json:"company" db:"company"`
}

// AddMetadata adds metadata to the job's tags.
func (j *Job) AddMetadata(key, value string) {
	if key == "" || value == "" {
		return
	}

	if j.Tags == nil {
		j.Tags = make(map[string][]string)
	}

	if key == "alternate_descriptions" {
		j.Tags[key] = append(j.Tags[key], value)
	} else {
		// split on comma and trim the string
		values := strings.Split(value, ",")
		for i := range values {
			values[i] = strings.TrimSpace(values[i])
		}

		j.Tags[key] = append(j.Tags[key], values...)
	}
	// deduplicate this key
	unique := make(map[string]struct{})
	for _, item := range j.Tags[key] {
		unique[item] = struct{}{}
	}

	j.Tags[key] = make([]string, 0, len(unique))
	for item := range unique {
		j.Tags[key] = append(j.Tags[key], item)
	}
}

// GetMetadata retrieves metadata values associated with a given key from the job's tags.
func (j *Job) GetMetadata(key string) []string {
	if j.Tags == nil {
		return nil
	}

	return j.Tags[key]
}

// GetSourceData retrieves the raw source data associated with the job.
func (j *Job) GetSourceData() []byte {
	return j.sourceData
}

// SetSourceData sets the raw source data for the job.
func (j *Job) SetSourceData(body []byte) {
	j.sourceData = body
}

// ProcessDatePosted processes and sets the DatePosted field from a JSON value.
func (j *Job) ProcessDatePosted(ctx context.Context, value []byte) {
	// let's see if this is an integer (unix timestamp)
	intValue, err := jsonparser.ParseInt(value)
	if err == nil {
		// assume this is a unix timestamp in seconds
		datePosted := time.Unix(intValue, 0).In(time.UTC)
		// let's make sure this is a date in the past
		if datePosted.Before(time.Now().In(time.UTC)) {
			// it is, set it
			j.DatePosted = datePosted
			return
		}

		slog.ErrorContext(ctx, "Parsed publishedDate is in the future", slog.Time("publishedDate", datePosted))
		// this is probably in milliseconds, try again
		datePosted = time.Unix(intValue/millisecondsToSeconds, 0).In(time.UTC)

		if datePosted.Before(time.Now().In(time.UTC)) {
			j.DatePosted = datePosted
			return
		}
	}

	// try parsing as a string
	stringValue, err := jsonparser.ParseString(value)
	if err != nil {
		slog.ErrorContext(ctx, "Error parsing publishedDate", slog.Any("error", err))
		return
	}

	datePosted, err := time.Parse("2006-01-02", stringValue)
	if err != nil {
		// if this is a time.ParseError, we can try to parse it as a full date-time
		datePosted, err = time.Parse(time.RFC3339, stringValue)
		if err == nil {
			j.DatePosted = datePosted.In(time.UTC) // Ensure the date is in UTC
		} else {
			slog.ErrorContext(ctx, "Error parsing publishedDate as date-time", slog.Any("error", err))
			return
		}
	} else {
		j.DatePosted = datePosted.In(time.UTC) // Ensure the date is in UTC
	}
}
