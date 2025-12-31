package bamboo

import (
	_ "embed"
	"testing"
)

//go:embed single_job.json
var singleJob string

func Test_parseBambooJob(t *testing.T) {
	t.Parallel()

	job, err := parseBambooJob(t.Context(), []byte(singleJob))
	if err != nil {
		t.Fatalf("parseBambooJob() error = %v", err)
	}

	if job.SourceID != "158" {
		t.Errorf("parseBambooJob() SourceID = %v, want %v", job.SourceID, "12345")
	}
}
