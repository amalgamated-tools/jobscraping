package models

import (
	"strings"
	"time"
)

type Job struct {
	URL              string         `json:"url"`
	CompensationUnit *string        `json:"compensation_unit"`
	DatePosted       time.Time      `json:"date_posted"`
	Department       Department     `json:"department"`
	Description      string         `json:"description"`
	EmploymentType   EmploymentType `json:"employment_type,omitempty"`
	Equity           EquityType     `json:"equity,omitempty"`
	IsRemote         bool           `json:"is_remote"`
	Location         string         `json:"location,omitempty"`
	LocationAddress  *string        `json:"location_address"`
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

func (j *Job) GetMetadata(key string) []string {
	if j.Tags == nil {
		return nil
	}

	return j.Tags[key]
}

func (j *Job) GetSourceData() []byte {
	return j.sourceData
}

func (j *Job) SetSourceData(body []byte) {
	j.sourceData = body
}
