// Package scheduler
package scheduler

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ljlericson/TaskForge/internal/job"
	"github.com/ljlericson/TaskForge/internal/registry"
)

type Queue interface {
	GetNextJob() (job.JobID, error)
	ReturnJobToQueue(job.JobID) error
	RemoveJobFromQueue(job.JobID) error
}

type Registry interface {
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
	queue           Queue
	registry        Registry
	logger          Logger
	activeTasks     map[job.JobID]*task // Tasks use same ID as jobs to avoid complication
	nodeIDToTask    map[registry.NodeID]job.JobID
	attemptsPerTask map[job.JobID]int
	mutex           sync.RWMutex
}

func NewScheduler(q Queue, r Registry, l Logger) *Scheduler {
	return &Scheduler{
		queue:        q,
		registry:     r,
		logger:       l,
		activeTasks:  make(map[job.JobID]*task),
		nodeIDToTask: make(map[registry.NodeID]job.JobID),
		mutex:        sync.RWMutex{},
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	s.logger.Infoln("starting scheduler")
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Infoln("shutting down scheduler")
			return
		case <-ticker.C:
			// first check if any of the active ndoes are dead
			s.mutex.RLock()
			tasks := make([]*task, 0, len(s.activeTasks))
			for _, t := range s.activeTasks {
				tasks = append(tasks, t)
			}
			s.mutex.RUnlock()

			for _, task := range tasks {
				if s.registry.IsNodeDead(task.NodeID) {
					s.JobFailed(task.JobID)
				}
			}
			// then check for new jobs and allocate jobs to free nodes
			if jobID, err := s.queue.GetNextJob(); err == nil {
				nodeID, err := s.registry.GetFreeNode()
				if err != nil {
					s.queue.ReturnJobToQueue(jobID)
					continue
				}

				s.mutex.Lock()
				_, ok := s.attemptsPerTask[jobID]
				if !ok {
					s.attemptsPerTask[jobID] = 0
				} else {
					s.attemptsPerTask[jobID]++
				}
				s.activeTasks[jobID] = &task{JobID: jobID, NodeID: nodeID, Attempt: s.attemptsPerTask[jobID]}
				s.nodeIDToTask[nodeID] = jobID
				s.mutex.Unlock()
			}
		}
	}
}

func (s *Scheduler) JobComplete(jobID job.JobID) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	_, ok := s.activeTasks[jobID]
	if !ok {
		return errors.New("job not found")
	}

	s.logger.Successln(fmt.Sprintf("job (ID: %s) SUCCEEDED, job removed from queue", string(jobID)))
	delete(s.activeTasks, jobID)
	delete(s.nodeIDToTask, s.activeTasks[jobID].NodeID)
	s.queue.RemoveJobFromQueue(jobID)
	return nil
}

func (s *Scheduler) JobFailed(jobID job.JobID) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
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

func (s *Scheduler) JobIDForNode(NodeID registry.NodeID) (job.JobID, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	jobID, ok := s.nodeIDToTask[NodeID]

	if !ok {
		return "", errors.New("node has no job asigned to it")
	}
	return jobID, nil
}
