package queue

import (
	"testing"

	"github.com/ljlericson/TaskForge/internal/heap"
	"github.com/ljlericson/TaskForge/internal/job"
	"github.com/ljlericson/TaskForge/internal/logging"
)

func setupQueue() *Queue {
	logger, _ := logging.NewLogger("logs/test.log")
	q := NewQueue(heap.NewHeap(), logger)

	jr := job.JobRequest{Priority: 1}
	j := job.Job{ID: "job1"}

	jr2 := job.JobRequest{Priority: 2}
	j2 := job.Job{ID: "job2"}

	jr3 := job.JobRequest{Priority: 3}
	j3 := job.Job{ID: "job3"}

	q.AddJobToQueue(&j3, &jr3)
	q.AddJobToQueue(&j, &jr)
	q.AddJobToQueue(&j2, &jr2)
	return q
}

func TestQueue_NextJobReturnsHighestPriority(t *testing.T) {
	q := setupQueue()
	j, err := q.GetNextJob()
	if err != nil {
		t.Fatal("expected queue to not give error")
	}

	if p, _ := q.GetJobPriority(j); p != 3 {
		t.Fatal("expected queue to return job with highest priority")
	}
}
