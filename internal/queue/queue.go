// Package queue
package queue

import (
	"sync"

	"github.com/ljlericson/TaskForge/internal/job"
)

var (
	Jobs     = make(chan job.Job, 1000)
	JobMap   = make(map[string]job.Job)
	NumJobs  int
	JobMutex = sync.RWMutex{}
)
