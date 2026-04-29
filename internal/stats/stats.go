package stats

import (
	"context"

	"github.com/ljlericson/TaskForge/internal/queue"
)

type Queue interface {
	GetJobs() []queue.JobInfo
}

type JobStatus struct {
	NodeID   string `json:"nodeID"`
	JobName  string `json:"jobName"`
	Progress int    `json:"progress"`
	Phase    int    `json:"phase"`
	Status   string `json:"status"`
}

type StatsPacket struct {
	Jobs  []JobStatus     `json:"activeJobs"`
	Queue []queue.JobInfo `json:"queue"`
	Nodes []string        `json:"nodes"`
}

type Stats struct {
	jobStatusReceived <-chan JobStatus
	jobStatusRequests chan chan StatsPacket
	jobStatusMap      map[string]JobStatus
	queue             Queue
}

func NewStats(jobStatusReceived <-chan JobStatus, q Queue) *Stats {
	return &Stats{
		jobStatusReceived: jobStatusReceived,
		jobStatusRequests: make(chan chan StatsPacket, 100),
		jobStatusMap:      map[string]JobStatus{},
		queue:             q,
	}
}

func (s *Stats) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case JobStatus := <-s.jobStatusReceived:
			s.jobStatusMap[JobStatus.JobName] = JobStatus
		case replyChan := <-s.jobStatusRequests:
			packet := s.buildStatsPacket()
			replyChan <- packet
		}
	}
}

func (s *Stats) RequestStatsPacket() StatsPacket {
	reply := make(chan StatsPacket)
	s.jobStatusRequests <- reply
	return <-reply
}

func (s *Stats) buildStatsPacket() StatsPacket {
	packet := StatsPacket{}

	for _, job := range s.jobStatusMap {
		packet.Jobs = append(packet.Jobs, job)
	}

	jobs := s.queue.GetJobs()
	packet.Queue = jobs

	return packet
}
