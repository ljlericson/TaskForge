package input

import (
	"bufio"
	"context"
	"os"
	"time"
)

type Handler func(string)

func Start(ctx context.Context, handler Handler) {

	scanner := bufio.NewScanner(os.Stdin)
	ticker := time.NewTicker(100 * time.Millisecond)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			text := scanner.Text()

			if handler != nil {
				go handler(text)
			}
		}
	}
}
