package input

import (
	"bufio"
	"context"
	"os"
	"time"
)

type Handler func(HandlerContext)

type Input struct {
	cancel  context.CancelFunc
	ctx     context.Context
	handler Handler
}

type HandlerContext struct {
	Cancel context.CancelFunc
	Input  string
}

func (i *Input) Start() {

	scanner := bufio.NewScanner(os.Stdin)
	ticker := time.NewTicker(100 * time.Millisecond)
	for scanner.Scan() {
		select {
		case <-i.ctx.Done():
			return

		case <-ticker.C:
			text := scanner.Text()

			if i.handler != nil {
				go i.handler(HandlerContext{
					Cancel: i.cancel,
					Input:  text,
				})
			}
		}
	}
}
