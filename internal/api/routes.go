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

func ConfigureRoutes(r *chi.Mux) {
	r.Get("/", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(LogoStr)) })
	r.Get("/jobs/next", JobNextHandler)
	r.Post("/jobs/submit", SubmitJobHandler)
	r.Post("/workers/register", SubmitJobHandler)
	r.Post("/workers/heatbeat", SubmitJobHandler)
}
