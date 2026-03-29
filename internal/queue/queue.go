// Package queue
package queue

import (
	"errors"
	"sync"

	"github.com/ljlericson/TaskForge/internal/heap"
	"github.com/ljlericson/TaskForge/internal/job"
)

type queueState struct {
	jobMap map[string]*job.Job
	reqMap map[string]*job.JobRequest
	mutex  sync.RWMutex
}

var queueStateInstance = queueState{
	jobMap: make(map[string]*job.Job),
	reqMap: make(map[string]*job.JobRequest),
	mutex:  sync.RWMutex{},
}

func AddJobToQueue(j *job.Job, jr *job.JobRequest) error {
	queueStateInstance.mutex.Lock()
	defer queueStateInstance.mutex.Unlock()
	if _, ok := queueStateInstance.jobMap[j.ID]; ok {
		return errors.New("job already exists")
	}

	err := heap.Push(j, jr)
	if err != nil {
		return err
	}

	queueStateInstance.jobMap[j.ID] = j
	queueStateInstance.reqMap[j.ID] = jr

	return nil
}

func GetNextJobReq() (*job.JobRequest, error) {
	queueStateInstance.mutex.RLock()
	defer queueStateInstance.mutex.RUnlock()

	key, err := heap.Pop()
	if err != nil {
		return nil, err
	}
	// console.C.Mutex.Lock()
	// console.C.Log(key)
	// console.C.Mutex.Unlock()

	job, ok := queueStateInstance.reqMap[key]
	if !ok {
		// should be impossible
		panic("job returned by heap does not exist")
	}

	return job, nil
}

func ReturnJobToQueue(ID string) error {
	queueStateInstance.mutex.Lock()
	defer queueStateInstance.mutex.Unlock()

	if _, ok := queueStateInstance.reqMap[ID]; !ok {
		return errors.New("job was not in queue")
	}

	err := heap.Push(queueStateInstance.jobMap[ID], queueStateInstance.reqMap[ID])
	return err
}

func RemoveJobFromQueue(ID string) error {
	queueStateInstance.mutex.Lock()
	defer queueStateInstance.mutex.Unlock()

	if _, ok := queueStateInstance.reqMap[ID]; !ok {
		return errors.New("job was not in queue")
	}

	delete(queueStateInstance.jobMap, ID)
	delete(queueStateInstance.reqMap, ID)
	return nil
}

func GetSizeOfQueue() int {
	queueStateInstance.mutex.RLock()
	defer queueStateInstance.mutex.RUnlock()
	return len(queueStateInstance.jobMap)
}
