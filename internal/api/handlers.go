// Package api
package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ljlericson/TaskForge/internal/job"
	"github.com/ljlericson/TaskForge/internal/queue"
)

func ListJobsHandler(w http.ResponseWriter, r *http.Request) {
	queue.JobMutex.RLock()
	for _, job := range queue.JobMap {
		responsStr := fmt.Sprintf("ID: %s\nTYPE: %s", job.ID, job.Type)
		w.Write([]byte(responsStr + "\n\n"))
	}
	queue.JobMutex.RUnlock()
}

func GetJobHabdler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	queue.JobMutex.Lock()
	j := queue.JobMap[id]
	queue.JobMutex.Unlock()
	responsStr := fmt.Sprintf("ID: %s\nTYPE: %s", j.ID, j.Type)
	w.Write([]byte(responsStr))
}

func SubmitJobHandler(w http.ResponseWriter, r *http.Request) {
	var j job.Job

	err := json.NewDecoder(r.Body).Decode(&j)
	if err != nil {
		w.Write([]byte(err.Error()))
	}

	queue.Jobs <- j

	queue.JobMutex.Lock()
	queue.JobMap[j.ID] = j
	queue.NumJobs++
	queue.JobMutex.Unlock()

	w.WriteHeader(200)
}
