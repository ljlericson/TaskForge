package app

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"

	"github.com/ljlericson/TaskForge/internal/api"
	"github.com/ljlericson/TaskForge/internal/input"
	"github.com/ljlericson/TaskForge/internal/logging"
)

type App struct {
	cfg    *Config
	ctx    context.Context
	cancel context.CancelFunc

	logger *logging.Logger
}

func NewApp(path string) (*App, error) {
	fmt.Print(logging.Yellow)
	fmt.Print(api.LogoStr)
	fmt.Println(logging.ResetCol)

	cfg, err := LoadConfig(path)
	if err != nil {
		return nil, err
	}

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
	)

	logger, err := logging.NewLogger(
		cfg.Logging.Path,
		cancel,
	)
	if err != nil {
		return nil, err
	}

	return &App{
		cfg:    cfg,
		ctx:    ctx,
		cancel: cancel,
		logger: logger,
	}, nil
}

func (a *App) Run() error {
	a.logger.Infoln("bootstrapping runtime")
	rt := a.Bootstrap()

	addr := net.JoinHostPort(a.cfg.Server.Host, strconv.Itoa(a.cfg.Server.Port))

	var wg sync.WaitGroup

	wg.Go(func() {
		if err := api.Server(a.ctx, addr, a.logger, rt.Handler); err != nil {
			a.logger.Abortln(err.Error())
		}
	})

	wg.Go(func() {
		rt.Scheduler.Start(a.ctx)
	})

	wg.Go(func() {
		rt.Registry.Start(a.ctx)
	})

	wg.Go(func() {
		rt.Input.Start(a.ctx)
	})

	<-a.ctx.Done()
	wg.Wait()

	return nil
}

func inputHandler(ctx input.HandlerContext) {
	args := strings.Fields(ctx.Input)
	if len(args) == 0 {
		return
	}

	switch args[0] {
	case "stop":
		ctx.Cancel()
	case "abort":
		ctx.Logger.Abortln("user triggered, reason: ", args[1:])
	case "health":
		ctx.HealthChecker.CheckHealth()
	}
}
