// Package queue
package queue

import (
	"errors"
	"sync"
	"time"

	"github.com/ljlericson/TaskForge/internal/job"
)

type Heap interface {
	Remove(id job.JobID) error
	NumberOfItems() int
	Pop() (job.JobID, error)
	Push(j *job.Job, priority int) error
}

type Logger interface {
	Infoln(v ...any)
	Warnln(v ...any)
	Successln(v ...any)
	Errorln(v ...any)
}

type JobInfo struct {
	JobName  string `json:"jobName"`
	Priority int    `json:"priority"`
}

type Queue struct {
	heap   Heap
	logger Logger
	jobMap map[job.JobID]*job.Job
	reqMap map[job.JobID]*job.JobRequest
	mutex  sync.RWMutex
}

func NewQueue(h Heap, l Logger) *Queue {
	return &Queue{
		heap:   h,
		logger: l,
		jobMap: make(map[job.JobID]*job.Job),
		reqMap: make(map[job.JobID]*job.JobRequest),
		mutex:  sync.RWMutex{},
	}
}

func (q *Queue) AddJobToQueue(jr *job.JobRequest) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	var j job.Job
	j.ID = job.JobID(jr.JobName)
	j.CreatedAt = time.Now()

	if _, ok := q.jobMap[j.ID]; ok {
		return errors.New("job already exists")
	}

	err := q.heap.Push(&j, jr.Priority)
	if err != nil {
		return err
	}

	q.jobMap[j.ID] = &j
	q.reqMap[j.ID] = jr

	return nil
}

func (q *Queue) GetNextJob() (job.JobID, error) {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	key, err := q.heap.Pop()
	if err != nil {
		return "", err
	}

	_, ok := q.reqMap[key]
	if !ok {
		// should be impossible
		panic("job returned by heap does not exist")
	}

	return key, nil
}

func (q *Queue) GetJobRequestFromID(ID job.JobID) (*job.JobRequest, error) {
	jr, ok := q.reqMap[ID]
	if !ok {
		return nil, errors.New("job (ID: " + string(ID) + ") does not exist")
	}
	return jr, nil
}

func (q *Queue) ReturnJobToQueue(ID job.JobID) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if _, ok := q.reqMap[ID]; !ok {
		return errors.New("job was not in queue")
	}

	err := q.heap.Push(q.jobMap[ID], q.reqMap[ID].Priority)
	return err
}

func (q *Queue) RemoveJobFromQueue(ID job.JobID) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if _, ok := q.reqMap[ID]; !ok {
		return errors.New("job was not in queue")
	}

	delete(q.jobMap, ID)
	delete(q.reqMap, ID)
	return nil
}

func (q *Queue) GetJobPriority(ID job.JobID) (int, error) {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	jr, ok := q.reqMap[ID]
	if !ok {
		return 0, errors.New("job does not exist")
	}
	return jr.Priority, nil
}

func (q *Queue) GetSizeOfQueue() int {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	return len(q.jobMap)
}
