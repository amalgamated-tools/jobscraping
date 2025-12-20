package ashby

import (
	"context"
	"fmt"
	"os"

	"github.com/amalgamated-tools/jobscraping/pkg/helpers"
)

var ashbyCompanyURL = "https://api.ashbyhq.com/posting-api/job-board/%s?includeCompensation=true"

func ScrapeCompany(ctx context.Context, company string) error {
	companyURL := fmt.Sprintf(ashbyCompanyURL, company)
	body, err := helpers.GetJSON(companyURL)
	if err != nil {
		return fmt.Errorf("error getting JSON from Ashby job board endpoint: %w", err)
	}

	fmt.Printf("Received response: %s\n", string(body))
	if err := os.WriteFile("examples/ashby/company.json", body, 0644); err != nil {
		return fmt.Errorf("error writing response to file: %w", err)
	}
	return nil
}
