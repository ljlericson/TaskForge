package input

import (
	"bufio"
	"context"
	"os"
)

type Handler func(string)

func Start(ctx context.Context, handler Handler) {

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {

		select {
		case <-ctx.Done():
			return

		default:
			text := scanner.Text()

			if handler != nil {
				go handler(text)
			}
		}
	}
}
