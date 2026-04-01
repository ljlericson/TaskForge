package heap

import (
	"testing"

	"github.com/ljlericson/TaskForge/internal/job"
)

func setupHeap() {
	jr := job.JobRequest{Priority: 1}
	j := job.Job{ID: "job1"}

	jr2 := job.JobRequest{Priority: 2}
	j2 := job.Job{ID: "job2"}

	jr3 := job.JobRequest{Priority: 3}
	j3 := job.Job{ID: "job3"}

	Push(&j3, &jr3)
	Push(&j, &jr)
	Push(&j2, &jr2)
}

func TestHeap_CorrectPriorityOrder(t *testing.T) {
	setupHeap()
	id1, err := Pop()
	id2, err2 := Pop()
	id3, err3 := Pop()

	if err != nil || err2 != nil || err3 != nil {
		t.Fatal("expected heap to not give error")
	}

	if id1 != "job3" || id2 != "job2" || id3 != "job1" {
		t.Fatal("expected heap to return jobs in order of highest priority")
	}
}

func TestHeap_DuplicateJob(t *testing.T) {
	setupHeap()

	jr := job.JobRequest{Priority: 1}
	j := job.Job{ID: "job1"}

	err := Push(&j, &jr)
	if err == nil {
		t.Fatal("expected heap to return error due to duplicate job")
	}
}
