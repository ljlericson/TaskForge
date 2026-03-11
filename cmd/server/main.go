package main

import (
	//	"encoding/json"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ljlericson/TaskForge/internal/console"
	"github.com/ljlericson/TaskForge/internal/job"
	"github.com/ljlericson/TaskForge/internal/queue"
)

const taskforgeLogo string = `
 _____ ___   _____ _   __   ______ ___________ _____  _____ 
|_   _/ _ \ /  ___| | / /   |  ___|  _  | ___ \  __ \|  ___|
  | |/ /_\ \\ '--.| |/ /    | |_  | | | | |_/ / |  \/| |__  
  | ||  _  | '--. \    \    |  _| | | | |    /| | __ |  __| 
  | || | | |/\__/ / |\  \   | |   \ \_/ / |\ \| |_\ \| |___ 
  \_/\_| |_/\____/\_| \_/   \_|    \___/\_| \_|\____/\____/ 

	`

func main() {
	fmt.Print(taskforgeLogo)

	time.Sleep(1 * time.Second)
	c := console.New()
	go server(c)

	time.Sleep(1 * time.Second)

	go func() {
		for cmd := range c.Input() {
			switch cmd {
			case "stop":
				c.Stop()
				return
			case "logo":
				c.Log(taskforgeLogo)
			default:
				c.Log("Command " + cmd + " is not a valid command")
			}
		}
	}()

	if err := c.Run(); err != nil {
		panic(err)
	}
}

func server(c *console.Console) {
	r := chi.NewRouter()
	r.Use(console.RequestLogger(c))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hi"))
	})
	r.Get("/jobs", listJobs)
	r.Get("/jobs/{id}", getJob)
	r.Post("/jobs", submitJob)
	http.ListenAndServe(":3000", r)
}

func listJobs(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("listing jobs"))
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func getJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	queue.JobMutex.Lock()
	j := queue.JobMap[id]
	queue.JobMutex.Unlock()
	responsStr := fmt.Sprintf("ID: %s\nTYPE: %s", j.ID, j.Type)
	w.Write([]byte(responsStr))
}

func submitJob(w http.ResponseWriter, r *http.Request) {
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
