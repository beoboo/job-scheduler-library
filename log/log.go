package log

import (
	"github.com/fatih/color"
	"os"
)

type Level int

const (
	Trace Level = 0
	Debug Level = 1
	Info  Level = 2
	Warn  Level = 3
	Error Level = 4
	Fatal Level = 5
)

type Mode int

const (
	Normal Mode = 0
	Dimmed Mode = 1
)

var logger Logger

type Logger struct {
	level Level
	mode  Mode
	trace *color.Color
	debug *color.Color
	info  *color.Color
	warn  *color.Color
	error *color.Color
	fatal *color.Color
}

func SetLevel(lvl Level) {
	logger.level = lvl
}

func SetMode(md Mode) {
	logger.mode = md
	switch md {
	case Dimmed:
		logger.trace = color.New(color.FgHiBlack)
		logger.debug = color.New(color.FgHiBlack)
		logger.info = color.New(color.FgHiBlack)
		logger.warn = color.New(color.FgHiBlack)
		logger.error = color.New(color.FgHiBlack)
		logger.fatal = color.New(color.FgHiBlack)
	default:
		logger.trace = color.New(color.FgBlue)
		logger.debug = color.New(color.FgHiBlack)
		logger.info = color.New(color.FgCyan)
		logger.warn = color.New(color.FgYellow)
		logger.error = color.New(color.FgRed)
		logger.fatal = color.New(color.FgHiRed)
	}
}

func init() {
	logger = Logger{
		level: Info,
	}

	SetMode(Normal)
}

func Traceln(args ...interface{}) {
	//fmt.Println(args...)
	if logger.level <= Trace {
		logger.trace.Println(args...)
	}
}

func Tracef(format string, args ...interface{}) {
	//fmt.Printf(format, args...)
	if logger.level <= Trace {
		logger.trace.Printf(format, args...)
	}
}

func Debugln(args ...interface{}) {
	//fmt.Println(args...)
	if logger.level <= Debug {
		logger.debug.Println(args...)
	}
}

func Debugf(format string, args ...interface{}) {
	//fmt.Printf(format, args...)
	if logger.level <= Debug {
		logger.debug.Printf(format, args...)
	}
}

func Infoln(args ...interface{}) {
	//fmt.Println(args...)
	if logger.level <= Info {
		logger.info.Println(args...)
	}
}

func Infof(format string, args ...interface{}) {
	//fmt.Printf(format, args...)
	if logger.level <= Info {
		logger.info.Printf(format, args...)
	}
}

func Warnln(args ...interface{}) {
	//fmt.Println(args...)
	if logger.level <= Warn {
		logger.warn.Println(args...)
	}
}

func Warnf(format string, args ...interface{}) {
	//fmt.Printf(format, args...)
	if logger.level <= Warn {
		logger.warn.Printf(format, args...)
	}
}

func Errorln(args ...interface{}) {
	//fmt.Println(args...)
	if logger.level <= Error {
		logger.error.Println(args...)
	}
}

func Errorf(format string, args ...interface{}) {
	//fmt.Printf(format, args...)
	if logger.level <= Error {
		logger.error.Printf(format, args...)
	}
}

func Fatalln(args ...interface{}) {
	//fmt.Println(args...)
	if logger.level <= Fatal {
		logger.fatal.Println(args...)
		os.Exit(1)
	}
}

func Fatalf(format string, args ...interface{}) {
	//fmt.Printf(format, args...)
	if logger.level <= Fatal {
		logger.fatal.Printf(format, args...)
		os.Exit(1)
	}
}

func Reset() {
	color.Unset()
}
