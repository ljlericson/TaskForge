// Package job
package job

type Job struct {
	ID        string
	Name      string
	Status    string
	Tasks     []*Task
	CreatedAt int64
}

type Task struct {
	ID          string
	JobID       string
	Input       string
	Output      string
	Status      string
	WorkerID    string
	Attempt     int
	LeaseExpiry int64
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
	Executors           int `json:"executors"`
	CoresPerExecutor    int `json:"coresPerExecutor"`
	MemoryPerExecutorMB int `json:"memoryPerExecutorMB"`
}

type DataSpec struct {
	Input  []string `json:"input"`
	Output string   `json:"output"`
}
