package console

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Console struct {
	app        *tview.Application
	logView    *tview.TextView
	input      *tview.InputField
	inputCh    chan string
	logBuffer  []string
	autoScroll bool

	Mutex   sync.RWMutex
	LogFile *os.File
}

var C *Console

func New(logFile *os.File) *Console {

	logView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true)

	logView.SetBorder(true).
		SetTitle(" TASKFORGE ")

	c := &Console{
		app:        tview.NewApplication(),
		logView:    logView,
		input:      tview.NewInputField(),
		inputCh:    make(chan string, 10),
		logBuffer:  []string{},
		autoScroll: true,
		LogFile:    logFile,
	}

	c.input.
		SetLabel("> ").
		SetFieldWidth(0)

	c.input.SetDoneFunc(func(key tcell.Key) {

		if key == tcell.KeyEnter {

			text := c.input.GetText()

			c.input.SetText("")

			c.inputCh <- text
		}
	})

	logView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		switch event.Key() {

		case tcell.KeyPgUp:
			row, col := logView.GetScrollOffset()

			pageSize := 20

			logView.ScrollTo(row-pageSize, col)
			c.autoScroll = false
			return nil

		case tcell.KeyPgDn:
			row, col := logView.GetScrollOffset()

			pageSize := 20

			logView.ScrollTo(row-pageSize, col)
			c.autoScroll = false
			return nil

		case tcell.KeyHome:
			logView.ScrollToBeginning()
			c.autoScroll = false
			return nil

		case tcell.KeyEnd:
			logView.ScrollToEnd()
			c.autoScroll = true
			return nil
		}

		return event
	})

	c.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		if event.Key() == tcell.KeyCtrlL {

			c.Clear()

			return nil
		}

		return event
	})

	layout :=
		tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(logView, 0, 1, false).
			AddItem(c.input, 1, 0, true)

	c.app.
		SetRoot(layout, true).
		EnableMouse(true)

	return c
}

func (c *Console) Clear() {

	c.Mutex.Lock()

	c.logBuffer = []string{}

	c.Mutex.Unlock()

	c.app.QueueUpdateDraw(func() {

		c.logView.Clear()
	})
}

func RequestLogger(c *Console) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			next.ServeHTTP(w, r)

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
		c.LogFile.Write([]byte(fmt.Sprintf("[%s] %s\n", timestamp, text)))
	})
}

func (c *Console) Input() <-chan string {
	return c.inputCh
}

func (c *Console) Run() error {
	return c.app.Run()
}

func (c *Console) Stop() {
	c.app.Stop()
}
