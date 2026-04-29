// Package api
package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

const LogoStr string = `
 _____ ___   _____ _   __   ______ ___________ _____  _____ 
|_   _/ _ \ /  ___| | / /   |  ___|  _  | ___ \  __ \|  ___|
  | |/ /_\ \\ '--.| |/ /    | |_  | | | | |_/ / |  \/| |__  
  | ||  _  | '--. \    \    |  _| | | | |    /| | __ |  __| 
  | || | | |/\__/ / |\  \   | |   \ \_/ / |\ \| |_\ \| |___ 
  \_/\_| |_/\____/\_| \_/   \_|    \___/\_| \_|\____/\____/ 

	`

func ConfigureRoutes(h *Handler, r *chi.Mux) {
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(LogoStr))
	})

	// public routes
	r.Post("/jobs/submit", h.SubmitJobHandler)

	// protected routes
	r.Group(func(r chi.Router) {
		r.Use(h.AuthMiddleware)

		r.Get("/jobs/next", h.NextJobHandler)
		r.Post("/jobs/status", h.JobStatusHandler)
		r.Post("/jobs/fail", h.JobFailHandler)
		r.Post("/workers/register", h.RegisterWorkerHandler)
		r.Post("/workers/heartbeat", h.WorkerHeartbeatHandler)
	})
}
