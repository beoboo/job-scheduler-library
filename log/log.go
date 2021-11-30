package log

import (
	"github.com/fatih/color"
	"os"
)

const (
	TRACE = 0
	DEBUG = 1
	INFO  = 2
	WARN  = 3
	ERROR = 4
	FATAL = 4
)

var logger Logger

type Logger struct {
	level int
	trace *color.Color
	debug *color.Color
	info  *color.Color
	warn  *color.Color
	error *color.Color
	fatal *color.Color
}

func init() {
	logger = Logger{
		level: DEBUG,
		trace: color.New(color.FgBlue),
		debug: color.New(color.FgHiBlack),
		info:  color.New(color.FgCyan),
		warn:  color.New(color.FgYellow),
		error: color.New(color.FgRed),
		fatal: color.New(color.FgHiRed),
	}
}

func Traceln(args ...interface{}) {
	if logger.level <= TRACE {
		logger.trace.Println(args...)
	}
}

func Tracef(format string, args ...interface{}) {
	if logger.level <= TRACE {
		logger.trace.Printf(format, args...)
	}
}

func Debugln(args ...interface{}) {
	if logger.level <= DEBUG {
		logger.debug.Println(args...)
	}
}

func Debugf(format string, args ...interface{}) {
	if logger.level <= DEBUG {
		logger.debug.Printf(format, args...)
	}
}

func Infoln(args ...interface{}) {
	if logger.level <= INFO {
		logger.info.Println(args...)
	}
}

func Infof(format string, args ...interface{}) {
	if logger.level <= INFO {
		logger.info.Printf(format, args...)
	}
}

func Warnln(args ...interface{}) {
	if logger.level <= WARN {
		logger.warn.Println(args...)
	}
}

func Warnf(format string, args ...interface{}) {
	if logger.level <= WARN {
		logger.warn.Printf(format, args...)
	}
}

func Errorln(args ...interface{}) {
	if logger.level <= ERROR {
		logger.error.Println(args...)
	}
}

func Errorf(format string, args ...interface{}) {
	if logger.level <= ERROR {
		logger.error.Printf(format, args...)
	}
}

func Fatalln(args ...interface{}) {
	if logger.level <= FATAL {
		logger.fatal.Println(args...)
		os.Exit(1)
	}
}

func Fatalf(format string, args ...interface{}) {
	if logger.level <= FATAL {
		logger.fatal.Printf(format, args...)
		os.Exit(1)
	}
}
