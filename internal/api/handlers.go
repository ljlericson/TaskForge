// Package api
package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/ljlericson/TaskForge/internal/job"
	"github.com/ljlericson/TaskForge/internal/queue"
	"github.com/ljlericson/TaskForge/internal/registry"
)

func JobNextHandler(w http.ResponseWriter, r *http.Request) {
	workerID := r.Header.Get("X-Worker-ID")
	sigHeader := r.Header.Get("X-Signature")
	timestamp := r.Header.Get("X-Timestamp")

	if workerID == "" || sigHeader == "" || timestamp == "" {
		http.Error(w, "missing auth headers", http.StatusBadRequest)
		return
	}

	sigBytes, err := base64.StdEncoding.DecodeString(sigHeader)
	if err != nil {
		http.Error(w, "invalid signature encoding", http.StatusBadRequest)
		return
	}

	tsInt, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		http.Error(w, "invalid timestamp", http.StatusBadRequest)
		return
	}

	now := time.Now().Unix()
	if abs(now-tsInt) > 30 {
		http.Error(w, "request expired", http.StatusUnauthorized)
		return
	}

	message := []byte(workerID + ":" + timestamp + ":" + r.Method + ":" + r.URL.Path)

	if !registry.AuthenticateWorker(workerID, message, sigBytes) {
		http.Error(w, "worker authentication failed", http.StatusUnauthorized)
	}

	job, err := queue.GetNextJobReq()
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err2 := json.NewEncoder(w).Encode(job)
	if err2 != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func SubmitJobHandler(w http.ResponseWriter, r *http.Request) {
	var jr job.JobRequest
	var j job.Job

	err := json.NewDecoder(r.Body).Decode(&jr)
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusBadRequest)
	}

	j.ID = jr.JobName
	j.CreatedAt = time.Now()

	err2 := queue.AddJobToQueue(&j, &jr)
	if err2 != nil {
		w.Write([]byte(err2.Error()))
		w.WriteHeader(http.StatusBadRequest)
	}

	w.WriteHeader(http.StatusOK)
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
