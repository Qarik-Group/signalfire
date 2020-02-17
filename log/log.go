package log

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

const (
	LevelError uint = iota
	LevelInfo
	LevelDebug
)

type Logger struct {
	Output io.Writer
	Level  uint
	lock   sync.Mutex
}

func (l *Logger) Fatal(f string, a ...interface{}) {
	l.write("FATAL", f, a...)
	os.Exit(1)
}

func (l *Logger) Error(f string, a ...interface{}) {
	l.write("ERROR", f, a...)
}

func (l *Logger) Info(f string, a ...interface{}) {
	if l.Level >= LevelInfo {
		l.write(" INFO", f, a...)
	}
}

func (l *Logger) Debug(f string, a ...interface{}) {
	if l.Level >= LevelDebug {
		l.write("DEBUG", f, a...)
	}
}

func (l *Logger) write(label, f string, a ...interface{}) {
	l.lock.Lock()
	fmt.Fprintf(l.Output, "%s: %s: "+f+"\n",
		append([]interface{}{time.Now().Format(time.RFC3339), label}, a...)...)
	l.lock.Unlock()
}
