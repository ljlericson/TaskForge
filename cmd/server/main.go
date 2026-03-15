package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/ljlericson/TaskForge/internal/api"
	"github.com/ljlericson/TaskForge/internal/console"
	"github.com/ljlericson/TaskForge/internal/job"
	"github.com/ljlericson/TaskForge/internal/queue"
	"github.com/ljlericson/TaskForge/internal/registry"
	"gopkg.in/yaml.v3"
)

func main() {
	config, err := loadConfig("config/server.yml")
	if err != nil {
		log.Fatalln(err)
		return
	}

	c := console.New(getLogFile(config))
	registry.InitRegistry(config.Server.Workers)

	go server(fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port), c)
	go processUserInput(c)

	if err := c.Run(); err != nil {
		panic(err)
	}
}

type config struct {
	Server  serverConfig  `yaml:"server"`
	Logging loggingConfig `yaml:"logging"`
	Session sessionConfig `yaml:"session"`
}

type serverConfig struct {
	Host    string                  `yaml:"host"`
	Port    int                     `yaml:"port"`
	Timeout int                     `yaml:"timeout"`
	Workers []registry.WorkerConfig `yaml:"workers"`
}

type loggingConfig struct {
	Path string `yaml:"path"`
}

type sessionConfig struct {
	Key string `yaml:"key"`
}

func server(addr string, c *console.Console) {
	r := chi.NewRouter()
	r.Use(console.RequestLogger(c))
	api.ConfigureRoutes(r)
	err := http.ListenAndServe(addr, r)
	if err != nil {
		panic(err)
	}
}

func processUserInput(c *console.Console) {
	for line := range c.Input() {

		args := strings.Fields(line)
		if len(args) == 0 {
			continue
		}

		switch args[0] {
		case "stop":
			c.Stop()
			return
		case "logo":
			c.Log(api.LogoStr)
		case "job":
			if len(args) == 3 {
				j := job.Job{}
				jr := job.JobRequest{}
				priority, err := strconv.Atoi(args[2])
				if err != nil {
					c.Log(err.Error())
					continue
				}
				j.ID = args[1]
				jr.Priority = priority
				err2 := queue.AddJobToQueue(&j, &jr)
				if err2 != nil {
					c.Log(err2.Error())
				}
			}
		case "getjob":
			job, err := queue.GetNextJobReq()
			if err != nil {
				c.Log(err.Error())
				continue
			}
			c.Log(fmt.Sprintf("%s : %d", job.JobName, job.Priority))
		default:
			c.Log("Command " + args[0] + " is not a valid command")
		}
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
