// Package queue
package queue

import (
	"errors"

	"github.com/ljlericson/TaskForge/internal/heap"
	"github.com/ljlericson/TaskForge/internal/job"
)

type queueState struct {
	jobMap map[string]*job.Job
	reqMap map[string]*job.JobRequest
}

var queueStateInstance = queueState{
	jobMap: make(map[string]*job.Job),
	reqMap: make(map[string]*job.JobRequest),
}

func AddJobToQueue(j *job.Job, jr *job.JobRequest) error {
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
	key, err := heap.Pop()
	if err != nil {
		return nil, err
	}

	job, ok := queueStateInstance.reqMap[key]
	if !ok {
		// should be impossible
		panic("job returned by heap does not exist")
	}

	return job, nil
}
