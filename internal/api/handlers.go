// Package api
package api

import (
	"encoding/json"
	"net/http"

	"github.com/ljlericson/TaskForge/internal/job"
)

func ListJobsHandler(w http.ResponseWriter, r *http.Request) {
}

func GetJobHabdler(w http.ResponseWriter, r *http.Request) {
}

func SubmitJobHandler(w http.ResponseWriter, r *http.Request) {
	var j job.Job

	err := json.NewDecoder(r.Body).Decode(&j)
	if err != nil {
		w.Write([]byte(err.Error()))
	}

	w.WriteHeader(200)
}
