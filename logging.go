package gts

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
)

var (
	red    = color.New(color.FgRed).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	blue   = color.New(color.BgBlue).SprintFunc()
)

type LogLevel int

const (
	SILENT LogLevel = iota
	ERROR
	WARN
	INFO
	DEBUG
)

func (level LogLevel) String() string {
	switch level {
	case ERROR:
		return red("ERROR")
	case WARN:
		return yellow(" WARN")
	case INFO:
		return green(" INFO")
	case DEBUG:
		return blue("DEBUG")
	default:
		panic(fmt.Sprintf("unknown log level: %d", level))
	}
}

var LOG_LEVEL = WARN

func SetLogLevel(level LogLevel) {
	LOG_LEVEL = level
}

var LOG_WRITER io.Writer = os.Stderr

func SetLogWriter(w io.Writer) {
	LOG_WRITER = w
}

func Logf(level LogLevel, format string, a ...interface{}) {
	if level <= LOG_LEVEL {
		msg := fmt.Sprintf(format, a...)
		fmt.Fprintf(LOG_WRITER, "[%s] %s", level.String(), msg)
	}
}

func Logln(level LogLevel, a ...interface{}) {
	if level <= LOG_LEVEL {
		msg := fmt.Sprintln(a...)
		fmt.Fprintf(LOG_WRITER, "[%s] %s", level.String(), msg)
	}
}

func Errorf(format string, a ...interface{}) {
	Logf(ERROR, format, a...)
}

func Errorln(a ...interface{}) {
	Logln(ERROR, a...)
}

func Warnf(format string, a ...interface{}) {
	Logf(WARN, format, a...)
}

func Warnln(a ...interface{}) {
	Logln(WARN, a...)
}

func Infof(format string, a ...interface{}) {
	Logf(INFO, format, a...)
}

func Infoln(a ...interface{}) {
	Logln(INFO, a...)
}

func Debugf(format string, a ...interface{}) {
	Logf(DEBUG, format, a...)
}

func Debugln(a ...interface{}) {
	Logln(DEBUG, a...)
}
