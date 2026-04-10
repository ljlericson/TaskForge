package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"

	"github.com/ljlericson/TaskForge/internal/api"
	"github.com/ljlericson/TaskForge/internal/input"
	"github.com/ljlericson/TaskForge/internal/job"
	"github.com/ljlericson/TaskForge/internal/logging"
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
	fmt.Print(logging.Yellow)
	fmt.Print(api.LogoStr)
	fmt.Println(logging.ResetCol)

	config, err := loadConfig("config/server.yml")
	if err != nil {
		log.Fatalln(err)
		return
	}

	addr := net.JoinHostPort(config.Server.Host, strconv.Itoa(config.Server.Port))

	logging.SetupLogger(config.Logging.Path)
	registry.InitRegistry(config.Workers)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	var wg sync.WaitGroup

	wg.Add(4)

	go func() {
		defer wg.Done()
		api.Server(ctx, addr)
	}()

	go func() {
		defer wg.Done()
		scheduler.Start(ctx)
	}()

	go func() {
		defer wg.Done()
		registry.CheckHeartbeats(ctx)
	}()

	go func() {
		defer wg.Done()
		input.Start(ctx, inputHandler)
	}()

	<-ctx.Done()
	wg.Wait()
}

func inputHandler(input string) {
	split := strings.Split(input, " ")
	for _, line := range split {
		args := strings.Fields(line)
		if len(args) == 0 {
			continue
		}

		switch args[0] {
		case "logo":
			logging.Infoln(api.LogoStr)
		case "job":
			if len(args) == 3 {
				j := job.Job{}
				jr := job.JobRequest{}
				priority, err := strconv.Atoi(args[2])
				if err != nil {
					logging.Errorln(err.Error())
					continue
				}
				j.ID = args[1]
				jr.Priority = priority
				err2 := queue.AddJobToQueue(&j, &jr)
				if err2 != nil {
					logging.Errorln(err2.Error())
				}
			}

		case "getjob":
			job, err := queue.GetNextJobReq()
			if err != nil {
				logging.Errorln(err.Error())
				continue
			}

			err2 := queue.RemoveJobFromQueue(job.JobName)
			if err2 != nil {
				logging.Errorln(err2.Error())
				continue
			}

			logging.Infoln(fmt.Sprintf("%s : %d", job.JobName, job.Priority))

		case "num":
			logging.Infoln(fmt.Sprintf("number of jobs: %d", queue.GetSizeOfQueue()))

		default:
			logging.Infoln("Command " + args[0] + " is not a valid command")
		}
	}
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
