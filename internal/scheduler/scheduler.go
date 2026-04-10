// Package schedular
package scheduler

import (
	"context"
	"errors"
	"sync"

	"github.com/ljlericson/TaskForge/internal/job"
	"github.com/ljlericson/TaskForge/internal/logging"
	"github.com/ljlericson/TaskForge/internal/queue"
	"github.com/ljlericson/TaskForge/internal/registry"
)

type schedulerState struct {
	jobsActive         map[string]*job.JobRequest // key is JobID
	nodeIDToActiveJobs map[string]*job.JobRequest // key is NodeID
	mutex              sync.RWMutex
}

var schedularStateInstance *schedulerState = &schedulerState{
	jobsActive:         make(map[string]*job.JobRequest),
	nodeIDToActiveJobs: make(map[string]*job.JobRequest),
	mutex:              sync.RWMutex{},
}

func Start(ctx context.Context) {
	logging.Infoln("starting schedular")
	for {
		select {
		case <-ctx.Done():
			logging.Infoln("shutting down schedular")
			return
		default:
			// first check for new jobs and allocate jobs to free nodes
			if jr, _ := queue.GetNextJobReq(); jr != nil {
				node, err := registry.GetFreeNode()
				if err != nil {
					queue.ReturnJobToQueue(jr.JobName)
					continue
				}
				schedularStateInstance.jobsActive[jr.JobName] = jr
				schedularStateInstance.nodeIDToActiveJobs[node.ID] = jr
			}
			// then check if any of the active ndoes are dead
			for NodeID, Jr := range schedularStateInstance.nodeIDToActiveJobs {
				if registry.IsNodeDead(NodeID) {
					JobFailed(Jr.JobName)
				}
			}
		}
	}
}

func JobComplete(JobID string) error {
	_, ok := schedularStateInstance.jobsActive[JobID]
	if !ok {
		return errors.New("job not found")
	}

	logging.Successln("job (ID: " + JobID + ") SUCCEEDED, job removed from queue")
	delete(schedularStateInstance.jobsActive, JobID)
	queue.RemoveJobFromQueue(JobID)
	return nil
}

func JobFailed(JobID string) error {
	_, ok := schedularStateInstance.jobsActive[JobID]
	if !ok {
		return errors.New("job not found")
	}

	logging.Warnln("job (ID: " + JobID + ") FAILED, returning to queue")

	delete(schedularStateInstance.jobsActive, JobID)
	queue.ReturnJobToQueue(JobID)
	return nil
}

func NodeDead(NodeID string) error {
	jr, ok := schedularStateInstance.nodeIDToActiveJobs[NodeID]
	if !ok {
		return errors.New("node does not have job asigned")
	}
	return JobFailed(jr.JobName)
}

func GetJobAssignedToNode(NodeID string) (*job.JobRequest, error) {
	schedularStateInstance.mutex.RLock()
	jr, ok := schedularStateInstance.nodeIDToActiveJobs[NodeID]
	schedularStateInstance.mutex.RUnlock()

	if !ok {
		return nil, errors.New("node has no job asigned to it")
	}
	return jr, nil
}
