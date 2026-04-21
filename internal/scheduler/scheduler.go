// Package scheduler
package scheduler

import (
	"context"
	"errors"
	"fmt"

	"github.com/ljlericson/TaskForge/internal/job"
	"github.com/ljlericson/TaskForge/internal/registry"
)

type jobRequest struct {
	nodeID registry.NodeID
	resp   chan job.JobID
}

type JobQueuer interface {
	GetNextJob() (job.JobID, error)
	AddJobToQueue(jr *job.JobRequest) error
	ReturnJobToQueue(job.JobID) error
	RemoveJobFromQueue(job.JobID) error
}

type Registry interface {
	GetIdleNodesChan() <-chan registry.NodeID
	GetDeadNodesChan() <-chan registry.NodeID
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
	jobSubmitted <-chan job.JobRequest
	nodeIdle     <-chan registry.NodeID
	nodeDead     <-chan registry.NodeID
	activeJobs   chan job.JobID
	errorChan    chan error
	jobCompleted chan job.JobID
	jobFailed    chan job.JobID
	jobRequests  chan jobRequest

	queue           JobQueuer
	registry        Registry
	logger          Logger
	activeTasks     map[job.JobID]*task // Tasks use same ID as jobs to avoid complication
	nodeIDToTask    map[registry.NodeID]job.JobID
	attemptsPerTask map[job.JobID]int
	idleNodes       []registry.NodeID
	activeNodes     map[registry.NodeID]struct{}
}

func NewScheduler(q JobQueuer, r Registry, l Logger, jobSumitted <-chan job.JobRequest) *Scheduler {
	return &Scheduler{
		jobSubmitted: jobSumitted,
		nodeIdle:     r.GetIdleNodesChan(),
		nodeDead:     r.GetDeadNodesChan(),
		activeJobs:   make(chan job.JobID, 100),
		errorChan:    make(chan error, 100),
		jobCompleted: make(chan job.JobID, 100),
		jobFailed:    make(chan job.JobID, 100),
		jobRequests:  make(chan jobRequest),

		queue:           q,
		registry:        r,
		logger:          l,
		activeTasks:     make(map[job.JobID]*task),
		nodeIDToTask:    make(map[registry.NodeID]job.JobID),
		idleNodes:       make([]registry.NodeID, 0),
		activeNodes:     make(map[registry.NodeID]struct{}),
		attemptsPerTask: map[job.JobID]int{},
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
				s.logger.Successln("dead node had no job asigned to it!")
				continue
			}
			s.failJob(jobID)

		case jr := <-s.jobSubmitted:
			s.errorChan <- s.queue.AddJobToQueue(&jr)

		case nodeID := <-s.nodeIdle:
			if _, ok := s.activeNodes[nodeID]; ok {
				previousJobID, ok := s.nodeIDToTask[nodeID]
				if !ok {
					s.logger.Errorln("active node gone idle did not have job asigned to it")
				} else {
					s.completeJob(previousJobID)
				}
			}
			s.logger.Infoln("node " + string(nodeID) + " is idle")
			s.idleNodes = append(s.idleNodes, nodeID)
			s.assignJobs()
		case req := <-s.jobRequests:
			// treat as idle node + assignment in one step

			if jobID, ok := s.nodeIDToTask[req.nodeID]; ok {
				// node was already running something → complete it
				s.completeJob(jobID)
			}

			jobID, err := s.queue.GetNextJob()
			if err != nil {
				req.resp <- "" // no job available
				continue
			}

			s.attemptsPerTask[jobID]++
			s.activeTasks[jobID] = &task{
				JobID:   jobID,
				NodeID:  req.nodeID,
				Attempt: s.attemptsPerTask[jobID],
			}

			s.nodeIDToTask[req.nodeID] = jobID
			s.activeNodes[req.nodeID] = struct{}{}

			s.logger.Infoln(fmt.Sprintf("job (ID: %s) assigned to node (ID: %s), job attempt %d", string(jobID), string(req.nodeID), s.attemptsPerTask[jobID]))

			req.resp <- jobID
		}
	}
}

func (s *Scheduler) RequestJob(nodeID registry.NodeID) (job.JobID, error) {
	resp := make(chan job.JobID)

	s.jobRequests <- jobRequest{
		nodeID: nodeID,
		resp:   resp,
	}

	jobID := <-resp
	if jobID == "" {
		return "", errors.New("no job available")
	}

	return jobID, nil
}

func (s *Scheduler) assignJobs() {
	for len(s.idleNodes) > 0 {
		jobID, err := s.queue.GetNextJob()
		if err != nil {
			// no more jobs in queue
			return
		}

		_, ok := s.attemptsPerTask[jobID]
		if !ok {
			s.attemptsPerTask[jobID] = 0
		}
		s.attemptsPerTask[jobID]++

		nodeID := s.idleNodes[0]
		s.idleNodes = s.idleNodes[1:]

		s.activeJobs <- jobID

		s.activeTasks[jobID] = &task{
			JobID:   jobID,
			NodeID:  nodeID,
			Attempt: s.attemptsPerTask[jobID],
		}
		s.nodeIDToTask[nodeID] = jobID
		s.activeNodes[nodeID] = struct{}{}
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
	if _, ok := s.activeNodes[task.NodeID]; !ok {
		s.logger.Errorln("no node found for task")
		return nil
	}

	delete(s.activeNodes, task.NodeID)
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

	if _, ok := s.activeNodes[task.NodeID]; !ok {
		s.logger.Errorln("no node found for task")
		return nil
	}
	delete(s.activeNodes, task.NodeID)
	delete(s.activeTasks, jobID)
	delete(s.nodeIDToTask, task.NodeID)

	return nil
}
