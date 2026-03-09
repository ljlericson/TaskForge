// Package console
package console

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Console represents our Paper-like server console.
type Console struct {
	app     *tview.Application
	logView *tview.TextView
	input   *tview.InputField
	inputCh chan string
	Mutex   sync.RWMutex
}

// New creates a new console.
func New() *Console {
	c := &Console{
		app:     tview.NewApplication(),
		logView: tview.NewTextView().SetDynamicColors(true).SetScrollable(true).SetChangedFunc(func() {}),
		input:   tview.NewInputField().SetLabel("> ").SetFieldWidth(0),
		inputCh: make(chan string, 10),
	}

	c.logView.SetBorder(true).SetTitle(" TASKFORGE ")
	c.input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			text := c.input.GetText()
			c.input.SetText("")
			c.inputCh <- text
		}
	})

	// Layout: logs on top, input on bottom
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(c.logView, 0, 1, false).
		AddItem(c.input, 1, 0, true)

	c.app.SetRoot(flex, true)
	return c
}

// RequestLogger returns a middleware that logs requests to the console
func RequestLogger(c *Console) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Call the next handler
			next.ServeHTTP(w, r)

			// Log the request to your console
			c.Log(
				r.Method + " " + r.URL.Path +
					" from " + r.RemoteAddr +
					" (" + time.Since(start).String() + ")",
			)
		})
	}
}

func (c *Console) Log(text string) {
	timestamp := time.Now().Format("29-10-2008 23:48:20")
	c.app.QueueUpdateDraw(func() {
		c.logView.Write([]byte(fmt.Sprintf("[%s] %s\n", timestamp, text)))
	})
}

// Input returns a channel that receives user-entered commands.
func (c *Console) Input() <-chan string {
	return c.inputCh
}

// Run starts the console UI. Blocks until application exits.
func (c *Console) Run() error {
	return c.app.Run()
}

// Stop stops the console UI.
func (c *Console) Stop() {
	c.app.Stop()
}
