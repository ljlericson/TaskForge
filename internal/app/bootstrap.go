package app

import (
	"github.com/ljlericson/TaskForge/internal/api"
	"github.com/ljlericson/TaskForge/internal/health"
	"github.com/ljlericson/TaskForge/internal/heap"
	"github.com/ljlericson/TaskForge/internal/input"
	"github.com/ljlericson/TaskForge/internal/job"
	"github.com/ljlericson/TaskForge/internal/queue"
	"github.com/ljlericson/TaskForge/internal/registry"
	"github.com/ljlericson/TaskForge/internal/scheduler"
	"github.com/ljlericson/TaskForge/internal/stats"
)

type Runtime struct {
	Registry  *registry.Registry
	Queue     *queue.Queue
	Scheduler *scheduler.Scheduler
	Handler   *api.Handler
	Input     *input.Input
	Events    *Events
	Health    *health.Health
}

type Events struct {
	JobSubmitted         chan job.JobRequest
	JobFailed            chan job.JobID
	NodeIdle             chan registry.NodeID
	NodeHeartbeat        chan registry.NodeID
	NodeRegistered       chan registry.NodeID
	JobStatusReceived    chan stats.JobStatus
	SchedulerCheckHealth chan chan struct{}
	RegistryCheckHealth  chan chan struct{}
}

func NewEvents() *Events {
	return &Events{
		JobSubmitted:         make(chan job.JobRequest, 100),
		JobFailed:            make(chan job.JobID, 100),
		NodeIdle:             make(chan registry.NodeID, 100),
		NodeHeartbeat:        make(chan registry.NodeID, 100),
		NodeRegistered:       make(chan registry.NodeID, 100),
		JobStatusReceived:    make(chan stats.JobStatus, 100),
		SchedulerCheckHealth: make(chan chan struct{}, 100),
		RegistryCheckHealth:  make(chan chan struct{}, 100),
	}
}

func (a *App) Bootstrap() *Runtime {
	events := NewEvents()

	reg := registry.NewRegistry(
		a.logger,
		events.NodeHeartbeat,
		events.NodeRegistered,
		events.RegistryCheckHealth,
	)

	reg.InitRegistry(a.cfg.Workers)

	h := heap.NewHeap()

	q := queue.NewQueue(
		h,
		a.logger,
	)

	sch := scheduler.NewScheduler(
		q,
		reg,
		a.logger,
		events.JobSubmitted,
		events.JobFailed,
		events.NodeIdle,
		events.SchedulerCheckHealth,
	)

	handler := api.NewHandler(
		sch,
		q,
		reg,
		a.logger,
		events.JobSubmitted,
		events.JobFailed,
		events.NodeRegistered,
		events.NodeIdle,
		events.NodeHeartbeat,
		events.JobStatusReceived,
	)

	hlth := health.NewHealth(
		events.SchedulerCheckHealth,
		events.RegistryCheckHealth,
		a.logger,
	)

	in := input.NewInput(
		a.cancel,
		hlth,
		a.logger,
	)

	in.AddHandler(inputHandler)

	return &Runtime{
		Registry:  reg,
		Queue:     q,
		Scheduler: sch,
		Handler:   handler,
		Input:     in,
		Events:    events,
		Health:    hlth,
	}
}
