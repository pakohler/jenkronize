package logging

import (
	"io"
	"io/ioutil"
	"log"
	"os"
)

var logger *Logger

type Logger struct {
	Flags       int
	Trace       *log.Logger
	Info        *log.Logger
	Warn        *log.Logger
	Error       *log.Logger
	Fatal       *log.Logger
	traceWriter io.Writer
	infoWriter  io.Writer
	warnWriter  io.Writer
	errorWriter io.Writer
	fatalWriter io.Writer
}

func (l *Logger) Init() *Logger {
	l.Trace = log.New(l.traceWriter, "TRACE:  ", l.Flags)
	l.Info = log.New(l.infoWriter, "INFO:  ", l.Flags)
	l.Warn = log.New(l.warnWriter, "WARN:  ", l.Flags)
	l.Error = log.New(l.errorWriter, "ERROR: ", l.Flags)
	l.Fatal = log.New(l.fatalWriter, "FATAL: ", l.Flags)
	return l
}

func (l *Logger) AddLogFile(logfile string) *Logger {
	f, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		l.Fatal.Fatal(err)
	}
	writers := []io.Writer{
		l.traceWriter,
		l.infoWriter,
		l.warnWriter,
		l.errorWriter,
		l.fatalWriter,
	}
	for i, w := range writers {
		if w != ioutil.Discard {
			writers[i] = io.MultiWriter(w, f)
		}
	}
	l.traceWriter = writers[0]
	l.infoWriter = writers[1]
	l.warnWriter = writers[2]
	l.errorWriter = writers[3]
	l.fatalWriter = writers[4]
	l.Init()
	return l
}

func GetLogger() *Logger {
	if logger == nil {
		logger = &Logger{
			Flags:       log.Ldate | log.Ltime | log.Lshortfile,
			traceWriter: ioutil.Discard,
			infoWriter:  os.Stdout,
			warnWriter:  os.Stdout,
			errorWriter: os.Stderr,
			fatalWriter: os.Stderr,
		}
		logger.Init()
	}
	return logger
}
