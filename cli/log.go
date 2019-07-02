package cli

import (
	"fmt"
	"os"
	"time"
)

// Log level
const (
	TRACE = iota
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
)

// ConsoleLoggerOption instance option
type ConsoleLoggerOption struct {
	level int
}

// ConsoleLogger type
type ConsoleLogger struct {
	prefix     string
	timeFormat string
	levels     []string
}

// Message type
type Message struct {
	level   int
	content string
	time    time.Time
}

// NewConsoleLogger constructor
func NewConsoleLogger() *ConsoleLogger {
	return &ConsoleLogger{
		prefix:     "",
		timeFormat: "15:04:05",
		levels:     []string{"TRACE", "DEBUG", "INFO ", "WARN ", "ERROR", "FATAL", "*LOG*"},
	}
}

func (l *ConsoleLogger) log(file *os.File, level int, format string, v ...interface{}) {
	msg := &Message{
		level:   level,
		content: fmt.Sprintf(format, v...),
		time:    time.Now(),
	}

	buf := []byte{}
	buf = append(buf, msg.time.Format(l.timeFormat)...)
	if l.prefix != "" {
		buf = append(buf, ' ')
		buf = append(buf, l.prefix...)
	}
	buf = append(buf, ' ')
	buf = append(buf, '[')
	buf = append(buf, l.levels[msg.level]...)
	buf = append(buf, ']')
	buf = append(buf, ' ')
	buf = append(buf, msg.content...)
	if len(msg.content) > 0 && msg.content[len(msg.content)-1] != '\n' {
		buf = append(buf, '\n')
	}
	file.Write(buf)
}

// Fatal fatal log
func (l *ConsoleLogger) Fatal(format string, v ...interface{}) {
	l.log(os.Stderr, FATAL, format, v...)
}

// Error fatal log
func (l *ConsoleLogger) Error(format string, v ...interface{}) {
	l.log(os.Stderr, ERROR, format, v...)
}

// Warn fatal log
func (l *ConsoleLogger) Warn(format string, v ...interface{}) {
	l.log(os.Stdout, WARN, format, v...)
}

// Debug fatal log
func (l *ConsoleLogger) Debug(format string, v ...interface{}) {
	l.log(os.Stdout, DEBUG, format, v...)
}

// Trace fatal log
func (l *ConsoleLogger) Trace(format string, v ...interface{}) {
	l.log(os.Stdout, TRACE, format, v...)
}

// Info fatal log
func (l *ConsoleLogger) Info(format string, v ...interface{}) {
	l.log(os.Stdout, INFO, format, v...)
}
