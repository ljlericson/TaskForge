// Package scheduler
package scheduler

import (
	"context"
	"errors"
	"fmt"

	"github.com/ljlericson/TaskForge/internal/job"
	"github.com/ljlericson/TaskForge/internal/registry"
)

type Queue interface {
	GetNextJob() (job.JobID, error)
	AddJobToQueue(jr *job.JobRequest) error
	ReturnJobToQueue(job.JobID) error
	RemoveJobFromQueue(job.JobID) error
}

type Registry interface {
	GetIdleNodesChan() <-chan registry.NodeID
	GetDeadNodesChan() <-chan registry.NodeID

	IsNodeDead(registry.NodeID) bool
	GetFreeNode() (registry.NodeID, error)
}

type Logger interface {
	Infoln(msg string)
	Warnln(msg string)
	Successln(msg string)
	Errorln(msg string)
}

type task struct {
	JobID   job.JobID
	Attempt int
	NodeID  registry.NodeID
}

type Scheduler struct {
	jobSubmited  <-chan job.JobRequest
	nodeIdle     <-chan registry.NodeID
	nodeDead     <-chan registry.NodeID
	activeJobs   chan job.JobID
	errorChan    chan error
	jobCompleted chan job.JobID
	jobFailed    chan job.JobID

	queue           Queue
	registry        Registry
	logger          Logger
	activeTasks     map[job.JobID]*task // Tasks use same ID as jobs to avoid complication
	nodeIDToTask    map[registry.NodeID]job.JobID
	attemptsPerTask map[job.JobID]int
	idleNodes       []registry.NodeID
}

func NewScheduler(q Queue, r Registry, l Logger, jobSumitted <-chan job.JobRequest) *Scheduler {
	return &Scheduler{
		jobSubmited: jobSumitted,
		nodeIdle:    r.GetIdleNodesChan(),
		nodeDead:    r.GetIdleNodesChan(),
		activeJobs:  make(chan job.JobID),

		queue:        q,
		registry:     r,
		logger:       l,
		activeTasks:  make(map[job.JobID]*task),
		nodeIDToTask: make(map[registry.NodeID]job.JobID),
		idleNodes:    make([]registry.NodeID, 0),
	}
}

func (s *Scheduler) GetActiveJobsChan() <-chan job.JobID {
	return s.activeJobs
}

func (s *Scheduler) GetErrorsChan() <-chan error {
	return s.errorChan
}

func (s *Scheduler) GetJobCompletedChan() <-chan job.JobID {
	return s.jobCompleted
}

func (s *Scheduler) GetJobFailedChan() <-chan job.JobID {
	return s.jobFailed
}

func (s *Scheduler) Start(ctx context.Context) {
	s.logger.Infoln("starting scheduler")

	for {
		select {
		case <-ctx.Done():
			s.logger.Infoln("shutting down scheduler")
			return
		case nodeID := <-s.nodeDead:
			jobID, ok := s.nodeIDToTask[nodeID]
			if !ok {
				// dead node was idle
				continue
			}
			s.failJob(jobID)

		case jr := <-s.jobSubmited:
			if err := s.queue.AddJobToQueue(&jr); err != nil {
				s.logger.Errorln("error adding job to queue " + err.Error())
				s.errorChan <- err
				continue
			}

		case nodeID := <-s.nodeIdle:
			s.idleNodes = append(s.idleNodes, nodeID)
			s.assignJobs()
		}
	}
}

func (s *Scheduler) assignJobs() {
	for len(s.idleNodes) > 0 {
		jobID, err := s.queue.GetNextJob()
		if err != nil {
			// no more jobs in queue
			return
		}

		nodeID := s.idleNodes[0]
		s.idleNodes = s.idleNodes[1:]

		s.activeJobs <- jobID
		s.logger.Infoln(fmt.Sprintf("job (ID: %s) assigned to node (ID: %s)", string(jobID), string(nodeID)))

		s.activeTasks[jobID] = &task{
			JobID:   jobID,
			NodeID:  nodeID,
			Attempt: s.attemptsPerTask[jobID],
		}
		s.nodeIDToTask[nodeID] = jobID
	}
}

func (s *Scheduler) completeJob(jobID job.JobID) error {
	task, ok := s.activeTasks[jobID]
	if !ok {
		return errors.New("job not found")
	}

	s.logger.Successln(fmt.Sprintf("job (ID: %s) SUCCEEDED, job removed from queue", string(jobID)))
	delete(s.activeTasks, jobID)
	delete(s.nodeIDToTask, task.NodeID)
	s.queue.RemoveJobFromQueue(jobID)
	return nil
}

func (s *Scheduler) failJob(jobID job.JobID) error {
	task, ok := s.activeTasks[jobID]
	if !ok {
		return errors.New("job not found")
	}

	s.logger.Warnln(fmt.Sprintf("job (ID: %s) FAILED, job returning to queue", string(jobID)))
	s.queue.ReturnJobToQueue(jobID)

	delete(s.activeTasks, jobID)
	delete(s.nodeIDToTask, task.NodeID)
	return nil
}

func (s *Scheduler) jobIDForNode(NodeID registry.NodeID) (job.JobID, error) {
	jobID, ok := s.nodeIDToTask[NodeID]

	if !ok {
		return "", errors.New("node has no job asigned to it")
	}
	return jobID, nil
}
