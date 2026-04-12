package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type MesCol string

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

	// bright foreground
	BrightBlack   MesCol = "\033[90m"
	BrightRed     MesCol = "\033[91m"
	BrightGreen   MesCol = "\033[92m"
	BrightYellow  MesCol = "\033[93m"
	BrightBlue    MesCol = "\033[94m"
	BrightMagenta MesCol = "\033[95m"
	BrightCyan    MesCol = "\033[96m"
	BrightWhite   MesCol = "\033[97m"

	// background
	BgBlack   MesCol = "\033[40m"
	BgRed     MesCol = "\033[41m"
	BgGreen   MesCol = "\033[42m"
	BgYellow  MesCol = "\033[43m"
	BgBlue    MesCol = "\033[44m"
	BgMagenta MesCol = "\033[45m"
	BgCyan    MesCol = "\033[46m"
	BgWhite   MesCol = "\033[47m"

	// bright background
	BgBrightBlack   MesCol = "\033[100m"
	BgBrightRed     MesCol = "\033[101m"
	BgBrightGreen   MesCol = "\033[102m"
	BgBrightYellow  MesCol = "\033[103m"
	BgBrightBlue    MesCol = "\033[104m"
	BgBrightMagenta MesCol = "\033[105m"
	BgBrightCyan    MesCol = "\033[106m"
	BgBrightWhite   MesCol = "\033[107m"
)

type Logger struct {
	fileLogger    *log.Logger
	consoleLogger *log.Logger
	mutex         sync.Mutex
}

func NewLogger(path string) (*Logger, error) {

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
	}, nil
}

func (l *Logger) log(col MesCol, level string, msg string) {

	l.mutex.Lock()
	defer l.mutex.Unlock()

	fmt.Print(col)
	l.consoleLogger.Printf("[%s] %s", level, msg)
	fmt.Print(ResetCol)

	l.fileLogger.Printf("[%s] %s", level, msg)
}

func (l *Logger) Infoln(msg string) {
	l.log(Cyan, "INFO", msg)
}

func (l *Logger) Warnln(msg string) {
	l.log(Yellow, "WARN", msg)
}

func (l *Logger) Successln(msg string) {
	l.log(Green, "SCCS", msg)
}

func (l *Logger) Errorln(msg string) {
	l.log(Red, "ERROR", msg)
}
