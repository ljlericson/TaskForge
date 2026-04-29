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
	"github.com/ljlericson/TaskForge/internal/heap"
	"github.com/ljlericson/TaskForge/internal/job"
	"github.com/ljlericson/TaskForge/internal/logging"
	"github.com/ljlericson/TaskForge/internal/queue"
	"github.com/ljlericson/TaskForge/internal/registry"
	"github.com/ljlericson/TaskForge/internal/scheduler"
	"github.com/ljlericson/TaskForge/internal/stats"
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
	}
	addr := net.JoinHostPort(config.Server.Host, strconv.Itoa(config.Server.Port))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	logger, err := logging.NewLogger(config.Logging.Path, stop)
	if err != nil {
		log.Fatalln(err)
	}

	jobSubmited := make(chan job.JobRequest, 100)
	jobFailed := make(chan job.JobID, 100)
	nodeIdle := make(chan registry.NodeID, 100)
	nodeHeartBeat := make(chan registry.NodeID, 100)
	nodeRegistered := make(chan registry.NodeID, 100)
	jobStatusRecieved := make(chan stats.JobStatus, 100)

	reg := registry.NewRegistry(logger, nodeHeartBeat, nodeRegistered)
	reg.InitRegistry(config.Workers)
	h := heap.NewHeap()
	q := queue.NewQueue(h, logger)
	sch := scheduler.NewScheduler(q, reg, logger, jobSubmited, jobFailed, nodeIdle)
	hand := api.NewHandler(sch, q, reg, logger, jobSubmited, jobFailed, nodeRegistered, nodeIdle, nodeHeartBeat, jobStatusRecieved)

	wg := sync.WaitGroup{}
	wg.Go(func() {
		if err := api.Server(ctx, addr, logger, hand); err != nil {
			logger.Abortln(err.Error())
		}
	})
	wg.Go(func() { sch.Start(ctx) })
	wg.Go(func() { reg.Start(ctx) })
	// wg.Go(func() { input.Start(ctx, stop, inputHandler) })

	<-ctx.Done()
	wg.Wait()
}

func inputHandler(cancel context.CancelFunc, input string) {
	for line := range strings.SplitSeq(input, " ") {
		args := strings.Fields(line)
		if len(args) == 0 {
			continue
		}

		switch args[0] {
		case "stop":
			cancel()
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
