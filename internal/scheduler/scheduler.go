// Package schedular
package scheduler

import (
	"context"

	"github.com/ljlericson/TaskForge/internal/job"
	"github.com/ljlericson/TaskForge/internal/queue"
	"github.com/ljlericson/TaskForge/internal/registry"
)

type schedulerState struct {
	nodesToAsignJobsMap map[string]*job.JobRequest
}

var schedularStateInstance *schedulerState = &schedulerState{
	nodesToAsignJobsMap: make(map[string]*job.JobRequest),
}

func Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if jr, _ := queue.GetNextJobReq(); jr != nil {
				node, err := registry.GetFreeNode()
				if err != nil {
					queue.ReturnJobToQueue(jr.JobName)
					continue
				}
				schedularStateInstance.nodesToAsignJobsMap[node.ID] = jr
			}
		}
	}
}
