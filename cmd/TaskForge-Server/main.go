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
	"github.com/ljlericson/TaskForge/internal/input"
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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	var wg sync.WaitGroup
	wg.Add(4)

	logger, err := logging.NewLogger(config.Logging.Path)
	if err != nil {
		panic(err)
	}

	var (
		registry  *registry.Registry   = registry.NewRegistry(logger)
		heap      *heap.Heap           = heap.NewHeap()
		queue     *queue.Queue         = queue.NewQueue(heap, logger)
		scheduler *scheduler.Scheduler = scheduler.NewScheduler(queue, registry, logger)
		handler   *api.Handler         = api.NewHandler(scheduler, queue, registry, logger)
	)

	go func() {
		defer wg.Done()
		api.Server(ctx, addr, logger, handler)
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
