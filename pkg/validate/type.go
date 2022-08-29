package validate

import (
	"gitlab.bit9.local/octarine/cbctl/pkg/model/resource"
)

// Job represents a validate job for specific resource that should be validated.
type Job struct {
	resourceData string
	filePath     string
	result       *resource.ValidatedResourceResponse
	error        string
}

// NewJob creates a new Job.
func NewJob(resourceData string, filePath string) *Job {
	return &Job{
		resourceData: resourceData,
		filePath:     filePath,
		result:       nil,
	}
}

// NewJobWithError creates a new Job with error.
func NewJobWithError(error string, filePath string) *Job {
	return &Job{
		error:    error,
		filePath: filePath,
		result:   nil,
	}
}
