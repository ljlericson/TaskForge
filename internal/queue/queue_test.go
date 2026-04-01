package queue

import (
	"testing"

	"github.com/ljlericson/TaskForge/internal/job"
)

func setupQueue() {
	jr := job.JobRequest{Priority: 1}
	j := job.Job{ID: "job1"}

	jr2 := job.JobRequest{Priority: 2}
	j2 := job.Job{ID: "job2"}

	jr3 := job.JobRequest{Priority: 3}
	j3 := job.Job{ID: "job3"}

	AddJobToQueue(&j3, &jr3)
	AddJobToQueue(&j, &jr)
	AddJobToQueue(&j2, &jr2)
}

func TestQueue_NextJobReturnsHighestPriority(t *testing.T) {
	setupQueue()
	jr, err := GetNextJobReq()
	if err != nil {
		t.Fatal("expected queue to not give error")
	}

	if jr.Priority != 3 {
		t.Fatal("expected queue to return job with highest priority")
	}
}
