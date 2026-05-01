package input

import (
	"bufio"
	"context"
	"os"
	"time"

	"github.com/ljlericson/TaskForge/internal/logging"
)

type Handler func(HandlerContext)
type Health interface {
	CheckHealth()
}

type Input struct {
	cancel   context.CancelFunc
	handlers []Handler
	health   Health
	logger   *logging.Logger
}

type HandlerContext struct {
	Cancel        context.CancelFunc
	Input         string
	HealthChecker Health
	Logger        *logging.Logger
}

func NewInput(cancel context.CancelFunc, health Health, logger *logging.Logger) *Input {
	return &Input{
		cancel:   cancel,
		handlers: make([]Handler, 0),
		health:   health,
		logger:   logger,
	}
}

func (i *Input) AddHandler(handler Handler) {
	if handler != nil {
		i.handlers = append(i.handlers, handler)
	}
}

func (i *Input) Start(ctx context.Context) {

	scanner := bufio.NewScanner(os.Stdin)
	ticker := time.NewTicker(100 * time.Millisecond)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			text := scanner.Text()

			for _, handler := range i.handlers {
				go handler(HandlerContext{
					Cancel:        i.cancel,
					Input:         text,
					Logger:        i.logger,
					HealthChecker: i.health,
				})
			}

		}
	}
}
