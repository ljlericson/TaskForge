// Package job
package job

import "time"

type JobStatus string
type JobID string

const (
	JobFinished JobStatus = "JobFinished"
	JobActive   JobStatus = "JobActive"
)

type Job struct {
	ID        JobID
	Status    JobStatus
	Progress  int
	CreatedAt time.Time
}

type JobRequest struct {
	JobName        string            `json:"jobName"`
	Jar            JarSpec           `json:"jar"`
	Resources      ResourceSpec      `json:"resources"`
	Data           DataSpec          `json:"data"`
	Arguments      []string          `json:"arguments"`
	Environment    map[string]string `json:"environment"`
	TimeoutSeconds int               `json:"timeoutSeconds"`
	Priority       int               `json:"priority"`
}

type JarSpec struct {
	URL       string `json:"url"`
	MainClass string `json:"mainClass"`
}

type ResourceSpec struct {
	ExecutionCores  int `json:"executionCores"`
	ExecutionMemory int `json:"executionMemory"`
}

type DataSpec struct {
	Input  []string `json:"input"`
	Output string   `json:"output"`
}
