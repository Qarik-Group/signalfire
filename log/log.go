package log

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

const (
	LevelFatal uint = iota
	LevelError
	LevelInfo
	LevelDebug
)

type Logger struct {
	Output io.Writer
	Level  uint
	lock   sync.Mutex
}

func (l *Logger) Fatal(f string, a ...interface{}) {
	l.write(LevelFatal, "FATAL", f, a...)
	os.Exit(1)
}

func (l *Logger) Error(f string, a ...interface{}) {
	l.write(LevelError, "ERROR", f, a...)
}

func (l *Logger) Info(f string, a ...interface{}) {
	l.write(LevelInfo, "_INFO", f, a...)
}

func (l *Logger) Debug(f string, a ...interface{}) {
	l.write(LevelDebug, "DEBUG", f, a...)
}

func (l *Logger) write(level uint, label, f string, a ...interface{}) {
	l.lock.Lock()
	if l.Level >= level {
		fmt.Fprintf(l.Output, "%s; %s; "+f+"\n",
			append([]interface{}{label, time.Now().Format(time.RFC3339)}, a...)...)
	}
	l.lock.Unlock()
}

func (l *Logger) SetLogLevel(level uint) {
	l.lock.Lock()
	l.Level = level
	l.lock.Unlock()
}
