package heap

import (
	"testing"

	"github.com/ljlericson/TaskForge/internal/job"
)

func setupHeap() *Heap {
	h := NewHeap()

	j := job.Job{ID: "job1"}

	j2 := job.Job{ID: "job2"}

	j3 := job.Job{ID: "job3"}

	h.Push(&j3, 3)
	h.Push(&j, 1)
	h.Push(&j2, 2)

	return h
}

func TestHeap_CorrectPriorityOrder(t *testing.T) {
	h := setupHeap()
	id1, err := h.Pop()
	id2, err2 := h.Pop()
	id3, err3 := h.Pop()

	if err != nil || err2 != nil || err3 != nil {
		t.Fatal("expected heap to not give error")
	}

	if id1 != "job3" || id2 != "job2" || id3 != "job1" {
		t.Fatal("expected heap to return jobs in order of highest priority")
	}
}

func TestHeap_DuplicateJob(t *testing.T) {
	h := setupHeap()

	j := job.Job{ID: "job1"}

	err := h.Push(&j, 1)
	if err == nil {
		t.Fatal("expected heap to return error due to duplicate job")
	}
}

func TestHeap_TopFunction(t *testing.T) {
	h := setupHeap()

	id, err := h.Top()
	if err != nil {
		t.Fatal("expected heap not return error")
	}
	if id != "job3" {
		t.Fatal("expected heap to return top but instead got " + id)
	}
	if h.NumberOfItems() != 3 {
		t.Fatal("expected heap size to remain constant")
	}
}

func TestHeap_PopOnEmptyHeap(t *testing.T) {
	h := NewHeap()
	_, err := h.Pop()

	if err == nil {
		t.Fatal("expected heap to give error due to empty heap")
	}
}

func TestHeap_TopOnEmptyHeap(t *testing.T) {
	h := NewHeap()
	_, err := h.Top()

	if err == nil {
		t.Fatal("expected heap to give error due to empty heap")
	}
}

func TestHeap_RemoveFunction(t *testing.T) {
	h := setupHeap()
	err := h.Remove("job3")

	if err != nil {
		t.Fatal("expected heap not return error")
	}
	if h.NumberOfItems() != 2 {
		t.Fatal("expected heap size to be 2")
	}
}

func TestHeap_RemoveGivenNonExistantJob(t *testing.T) {
	h := setupHeap()
	err := h.Remove("job4") // job doesn't exist

	if err == nil {
		t.Fatal("expected heap to return error")
	}
}
