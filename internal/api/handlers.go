// Package api
package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/ljlericson/TaskForge/internal/job"
	"github.com/ljlericson/TaskForge/internal/logging"
	"github.com/ljlericson/TaskForge/internal/queue"
	"github.com/ljlericson/TaskForge/internal/registry"
	"github.com/ljlericson/TaskForge/internal/scheduler"
)

type ctxKey int

const (
	headerWorkerID  = "X-Worker-ID"
	headerSignature = "X-Signature"
	headerTimestamp = "X-Timestamp"

	maxClockSkew = 5 * time.Minute
	maxBodySize  = 1 << 20 // 1 MB

	workerIDKey ctxKey = iota
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

		workerID := r.Header.Get(headerWorkerID)
		sigHeader := r.Header.Get(headerSignature)
		timestamp := r.Header.Get(headerTimestamp)

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

		if time.Since(time.Unix(tsInt, 0)) > maxClockSkew {
			http.Error(w, "request expired", http.StatusUnauthorized)
			return
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}
		r.Body.Close()

		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		hash := sha256.Sum256(bodyBytes)

		message := []byte(
			workerID + ":" +
				timestamp + ":" +
				r.Method + ":" +
				r.URL.Path + ":" +
				hex.EncodeToString(hash[:]),
		)

		if !registry.AuthenticateWorker(workerID, message, sigBytes) {
			http.Error(w, "worker authentication failed", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), workerIDKey, workerID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func WorkerIDFromContext(ctx context.Context) string {
	v, ok := ctx.Value(workerIDKey).(string)
	if !ok {
		return ""
	}
	return v
}

func NextJobHandler(w http.ResponseWriter, r *http.Request) {
	workerID := WorkerIDFromContext(r.Context())

	jr, err := scheduler.GetJobAssignedToNode(workerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(jr); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func SubmitJobHandler(w http.ResponseWriter, r *http.Request) {
	var jr job.JobRequest
	var j job.Job

	err := json.NewDecoder(r.Body).Decode(&jr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	j.ID = jr.JobName
	j.CreatedAt = time.Now()

	if err := queue.AddJobToQueue(&j, &jr); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	logging.Infoln("new job submitted (ID: " + jr.JobName + ")")
}

func RegisterWorkerHandler(w http.ResponseWriter, r *http.Request) {
	var newNode registry.Node

	if err := json.NewDecoder(r.Body).Decode(&newNode); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	newNode.Status = registry.NodeHealthy

	if err := registry.RegisterNode(&newNode); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	logging.Infoln("worker (ID: " + newNode.ID + ") has registered successfully")
}

func WorkerHeartbeatHandler(w http.ResponseWriter, r *http.Request) {
	workerID := WorkerIDFromContext(r.Context())
	var heartbeat registry.Heatbeat

	if err := json.NewEncoder(w).Encode(&heartbeat); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	heartbeat.ID = workerID

	if err := registry.RegisterHeatbeat(heartbeat); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
