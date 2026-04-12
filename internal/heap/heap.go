// Package heap
package heap

import (
	"container/heap"
	"errors"
	"sync"

	"github.com/ljlericson/TaskForge/internal/job"
)

type heapItem struct {
	id       job.JobID
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

type Heap struct {
	pq    priorityQueue
	items map[job.JobID]*heapItem
	mutex sync.RWMutex
}

func NewHeap() *Heap {
	return &Heap{
		pq:    make(priorityQueue, 0),
		items: make(map[job.JobID]*heapItem),
		mutex: sync.RWMutex{},
	}
}

func (h *Heap) Push(j *job.Job, priority int) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	if _, ok := h.items[j.ID]; ok {
		return errors.New("job already exists")
	}

	item := &heapItem{
		id:       j.ID,
		priority: priority,
	}

	heap.Push(&h.pq, item)
	h.items[j.ID] = item

	return nil
}

func (h *Heap) Top() (job.JobID, error) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	if h.pq.Len() == 0 {
		return "", errors.New("heap empty")
	}

	return h.pq[0].id, nil
}

func (h *Heap) Pop() (job.JobID, error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	if h.pq.Len() == 0 {
		return "", errors.New("heap empty")
	}

	item := heap.Pop(&h.pq).(*heapItem)

	delete(h.items, item.id)

	return item.id, nil
}

func (h *Heap) Remove(id job.JobID) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	item, ok := h.items[id]
	if !ok {
		return errors.New("job does not exist")
	}

	heap.Remove(&h.pq, item.index)

	delete(h.items, id)

	return nil
}

func (h *Heap) NumberOfItems() int {
	return len(h.items)
}
