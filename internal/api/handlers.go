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
	"github.com/ljlericson/TaskForge/internal/stats"
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

type Handler struct {
	jobSubmitted      chan job.JobRequest
	jobFailed         chan job.JobID
	jobStatusRecieved chan stats.JobStatus
	nodeIdle          chan registry.NodeID
	nodeHeartBeat     chan registry.NodeID
	nodeRegistered    chan registry.NodeID
	schedulerErrChan  <-chan error
	registryErrChan   <-chan error

	scheduler *scheduler.Scheduler
	queue     *queue.Queue
	registry  *registry.Registry
	logger    *logging.Logger
}

func NewHandler(s *scheduler.Scheduler,
	q *queue.Queue,
	r *registry.Registry,
	l *logging.Logger,
	jobSubmitted chan job.JobRequest,
	jobFailed chan job.JobID,
	nodeRegistered chan registry.NodeID,
	nodeIdle chan registry.NodeID,
	nodeHeartBeat chan registry.NodeID,
	jobStatusRecieved chan stats.JobStatus,
) *Handler {
	return &Handler{
		jobSubmitted:      jobSubmitted,
		jobFailed:         jobFailed,
		jobStatusRecieved: jobStatusRecieved,
		nodeRegistered:    nodeRegistered,
		nodeIdle:          nodeIdle,
		nodeHeartBeat:     nodeHeartBeat,
		schedulerErrChan:  s.GetErrorsChan(),
		registryErrChan:   r.GetErrorChan(),
		scheduler:         s,
		queue:             q,
		registry:          r,
		logger:            l,
	}
}

func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
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

		if err := h.registry.AuthenticateWorker(workerID, message, sigBytes); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), workerIDKey, workerID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func workerIDFromContext(ctx context.Context) string {
	v, ok := ctx.Value(workerIDKey).(string)
	if !ok {
		return ""
	}
	return v
}

func (h *Handler) NextJobHandler(w http.ResponseWriter, r *http.Request) {
	workerID := workerIDFromContext(r.Context())

	h.nodeIdle <- registry.NodeID(workerID)

	jobID, err := h.scheduler.RequestJob(registry.NodeID(workerID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNoContent)
		return
	}

	jr, err := h.queue.GetJobRequestFromID(jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(jr); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) SubmitJobHandler(w http.ResponseWriter, r *http.Request) {
	var jr job.JobRequest

	if err := json.NewDecoder(r.Body).Decode(&jr); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.logger.Infoln("handler sent signal on job submit channel")
	h.jobSubmitted <- jr

	select {
	case err := <-h.schedulerErrChan:
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	default:
	}

	h.logger.Infoln("new job submitted (ID: " + jr.JobName + ")")
}

func (h *Handler) JobStatusHandler(w http.ResponseWriter, r *http.Request) {
	var js stats.JobStatus

	if err := json.NewDecoder(r.Body).Decode(&js); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.jobStatusRecieved <- js
}

func (h *Handler) JobFailHandler(w http.ResponseWriter, r *http.Request) {
	var j struct {
		JobID string `json:"jobID"`
	}

	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.logger.Infoln("handler sent signal on job fail channel")
	h.jobFailed <- job.JobID(j.JobID)

	select {
	case err := <-h.schedulerErrChan:
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			h.logger.Errorln(err.Error())
			return
		}
	default:
	}
}

func (h *Handler) RegisterWorkerHandler(w http.ResponseWriter, r *http.Request) {
	workerID := workerIDFromContext(r.Context())

	h.logger.Infoln("handler sent signal on node registered channel")
	h.nodeRegistered <- registry.NodeID(workerID)
	select {
	case err := <-h.registryErrChan:
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	default:
	}
}

func (h *Handler) WorkerHeartbeatHandler(w http.ResponseWriter, r *http.Request) {
	workerID := workerIDFromContext(r.Context())

	h.nodeHeartBeat <- registry.NodeID(workerID)

	select {
	case err := <-h.registryErrChan:
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	default:
	}
}
