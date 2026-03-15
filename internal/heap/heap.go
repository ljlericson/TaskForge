// Package heap
package heap

import (
	"container/heap"
	"errors"

	"github.com/ljlericson/TaskForge/internal/job"
)

type heapItem struct {
	id       string
	priority int
	index    int
}

type priorityQueue []*heapItem

func (pq priorityQueue) Len() int           { return len(pq) }
func (pq priorityQueue) Less(i, j int) bool { return pq[i].priority > pq[j].priority }
func (pq priorityQueue) Swap(i, j int)      { pq[i], pq[j] = pq[j], pq[i]; pq[i].index = i; pq[j].index = j }

func (pq *priorityQueue) Push(x any) {
	item := x.(*heapItem)
	item.index = len(*pq)
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() any {
	old := *pq
	n := len(old)

	item := old[n-1]
	old[n-1] = nil
	item.index = -1

	*pq = old[0 : n-1]

	return item
}

type heapState struct {
	pq     priorityQueue
	items  map[string]*heapItem
	jobMap map[string]*job.Job
	reqMap map[string]*job.JobRequest
}

var heapStateInstance = heapState{
	pq:     make(priorityQueue, 0),
	items:  make(map[string]*heapItem),
	jobMap: make(map[string]*job.Job),
	reqMap: make(map[string]*job.JobRequest),
}

func Push(j *job.Job, req *job.JobRequest) error {
	if _, ok := heapStateInstance.items[j.ID]; ok {
		return errors.New("job already exists")
	}

	item := &heapItem{
		id:       j.ID,
		priority: req.Priority,
	}

	heap.Push(&heapStateInstance.pq, item)

	heapStateInstance.items[j.ID] = item
	heapStateInstance.jobMap[j.ID] = j
	heapStateInstance.reqMap[j.ID] = req

	return nil
}

func Top() (string, error) {
	if heapStateInstance.pq.Len() == 0 {
		return "", errors.New("heap empty")
	}

	return heapStateInstance.pq[0].id, nil
}

func Pop() (string, error) {
	if heapStateInstance.pq.Len() == 0 {
		return "", errors.New("heap empty")
	}

	item := heap.Pop(&heapStateInstance.pq).(*heapItem)

	delete(heapStateInstance.items, item.id)
	delete(heapStateInstance.jobMap, item.id)
	delete(heapStateInstance.reqMap, item.id)

	return item.id, nil
}

func Remove(id string) error {
	item, ok := heapStateInstance.items[id]
	if !ok {
		return errors.New("job does not exist")
	}

	heap.Remove(&heapStateInstance.pq, item.index)

	delete(heapStateInstance.items, id)
	delete(heapStateInstance.jobMap, id)
	delete(heapStateInstance.reqMap, id)

	return nil
}
