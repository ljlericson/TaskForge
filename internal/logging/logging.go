package logging

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type MesCol string
type BackgroundCol string

const (
	// reset
	ResetCol MesCol = "\033[0m"

	// styles
	Bold      MesCol = "\033[1m"
	Dim       MesCol = "\033[2m"
	Italic    MesCol = "\033[3m"
	Underline MesCol = "\033[4m"
	Blink     MesCol = "\033[5m"
	Reverse   MesCol = "\033[7m"
	Hidden    MesCol = "\033[8m"
	Strike    MesCol = "\033[9m"

	// standard foreground
	Black   MesCol = "\033[30m"
	Red     MesCol = "\033[31m"
	Green   MesCol = "\033[32m"
	Yellow  MesCol = "\033[33m"
	Blue    MesCol = "\033[34m"
	Magenta MesCol = "\033[35m"
	Cyan    MesCol = "\033[36m"
	White   MesCol = "\033[37m"

	// background
	BgBlack   BackgroundCol = "\033[40m"
	BgRed     BackgroundCol = "\033[41m"
	BgGreen   BackgroundCol = "\033[42m"
	BgYellow  BackgroundCol = "\033[43m"
	BgBlue    BackgroundCol = "\033[44m"
	BgMagenta BackgroundCol = "\033[45m"
	BgCyan    BackgroundCol = "\033[46m"
	BgWhite   BackgroundCol = "\033[47m"
)

type Logger struct {
	fileLogger    *log.Logger
	consoleLogger *log.Logger
	mutex         sync.Mutex
	cancel        context.CancelFunc
}

func NewLogger(path string, cancel context.CancelFunc) (*Logger, error) {

	logDir := filepath.Dir(path)
	err := os.MkdirAll(logDir, os.ModePerm)
	if err != nil {
		panic(err)
	}

	file, err2 := os.Create(path)
	if err2 != nil {
		panic(err2)
	}

	return &Logger{
		fileLogger:    log.New(file, "", log.LstdFlags),
		consoleLogger: log.New(os.Stdout, "", log.LstdFlags),
		cancel:        cancel,
	}, nil
}

func (l *Logger) log(col MesCol, prefixCol BackgroundCol, level string, v ...any) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	prefix := "[" + level + "]"
	msg := fmt.Sprint(v...)

	var b strings.Builder

	if prefixCol == "" {
		prefixCol = BackgroundCol(string(col))
	}
	b.WriteString(string(Bold))
	b.WriteString(string(prefixCol))
	b.WriteString(prefix)
	b.WriteString(string(ResetCol))
	b.WriteRune(' ')
	b.WriteString(string(col))
	b.WriteString(msg)
	b.WriteString(string(ResetCol))

	l.consoleLogger.Println(b.String())
	l.fileLogger.Println(append([]any{prefix}, v...)...)
}

func (l *Logger) Infoln(v ...any) {
	l.log(Cyan, "", "INFO", v...)
}

func (l *Logger) Warnln(v ...any) {
	l.log(Yellow, "", "WARN", v...)
}

func (l *Logger) Successln(v ...any) {
	l.log(Green, BgGreen, "OK", v...)
}

func (l *Logger) Errorln(v ...any) {
	l.log(Red, BgRed, "ERROR", v...)
}

func (l *Logger) Abortln(v ...any) {
	l.log(Red, BgMagenta, "ABORT", v...)
	l.cancel()
}

func (l *Logger) Fatalln(v ...any) {
	l.log(Red, BgMagenta, "FATAL", v...)
	os.Exit(1)
}
