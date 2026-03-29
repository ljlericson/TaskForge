package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/ljlericson/TaskForge/internal/api"
	"github.com/ljlericson/TaskForge/internal/console"
	"github.com/ljlericson/TaskForge/internal/job"
	"github.com/ljlericson/TaskForge/internal/queue"
	"github.com/ljlericson/TaskForge/internal/registry"
	"github.com/ljlericson/TaskForge/internal/scheduler"
	"gopkg.in/yaml.v3"
)

type config struct {
	Server  serverConfig            `yaml:"server"`
	Logging loggingConfig           `yaml:"logging"`
	Session sessionConfig           `yaml:"session"`
	Workers []registry.WorkerConfig `yaml:"workers"`
}

type serverConfig struct {
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	Timeout int    `yaml:"timeout"`
}

type loggingConfig struct {
	Path string `yaml:"path"`
}

type sessionConfig struct {
	Key string `yaml:"key"`
}

func main() {
	config, err := loadConfig("config/server.yml")
	if err != nil {
		log.Fatalln(err)
		return
	}

	console.C = console.New(getLogFile(config))
	registry.InitRegistry(config.Workers)
	ctx, cancel := context.WithCancel(context.Background())

	go server(ctx, fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port))
	go scheduler.Start(ctx)
	go registry.CheckHeartbeats(ctx)
	go processUserInput(ctx, cancel)

	if err := console.C.Run(); err != nil {
		panic(err)
	}
}

func server(ctx context.Context, addr string) error {

	r := chi.NewRouter()

	r.Use(console.RequestLogger(console.C))

	r.Use(cors.Handler(cors.Options{

		AllowedOrigins: []string{"*"},

		AllowedMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"OPTIONS",
		},

		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
		},

		AllowCredentials: false,
		MaxAge:           300,
	}))

	api.ConfigureRoutes(r)

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {

		if err := srv.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {

			panic(err)
		}

	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	return srv.Shutdown(shutdownCtx)
}

func processUserInput(ctx context.Context, cancel context.CancelFunc) {
	for line := range console.C.Input() {
		console.C.Mutex.Lock()
		select {
		case <-ctx.Done():
			return
		default:
			args := strings.Fields(line)
			if len(args) == 0 {
				continue
			}

			switch args[0] {
			case "stop":
				cancel()
				console.C.Stop()
				return
			case "logo":
				console.C.Log(api.LogoStr)
			case "job":
				if len(args) == 3 {
					j := job.Job{}
					jr := job.JobRequest{}
					priority, err := strconv.Atoi(args[2])
					if err != nil {
						console.C.Log(err.Error())
						continue
					}
					j.ID = args[1]
					jr.Priority = priority
					err2 := queue.AddJobToQueue(&j, &jr)
					if err2 != nil {
						console.C.Log(err2.Error())
					}
				}
			case "getjob":
				job, err := queue.GetNextJobReq()
				if err != nil {
					console.C.Log(err.Error())
					continue
				}
				console.C.Log(fmt.Sprintf("%s : %d", job.JobName, job.Priority))
			case "num":
				console.C.Log(fmt.Sprintf("number of jobs: %d", queue.GetSizeOfQueue()))
			default:
				console.C.Log("Command " + args[0] + " is not a valid command")
			}
		}
		console.C.Mutex.Unlock()
		time.Sleep(1 * time.Second)
	}
}

func getLogFile(config *config) *os.File {
	logDir := filepath.Dir(config.Logging.Path)
	err3 := os.MkdirAll(logDir, os.ModePerm)
	if err3 != nil {
		panic(err3)
	}

	logFile, err2 := os.Create(config.Logging.Path)
	if err2 != nil {
		panic(err2)
	}
	return logFile
}

func loadConfig(path string) (*config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg config

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
