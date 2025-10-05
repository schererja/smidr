package source

import (
	"fmt"
	"io"
	"time"
)

type ConsoleLogger struct {
	output  io.Writer
	verbose bool
}

func NewConsoleLogger(output io.Writer, verbose bool) *ConsoleLogger {
	return &ConsoleLogger{
		output:  output,
		verbose: verbose,
	}
}

func (l *ConsoleLogger) Info(msg string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	fmt.Fprintf(l.output, "[%s] INFO: %s\n", timestamp, fmt.Sprintf(msg, args...))
}

func (l *ConsoleLogger) Error(msg string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	fmt.Fprintf(l.output, "[%s] ERROR: %s\n", timestamp, fmt.Sprintf(msg, args...))
}

func (l *ConsoleLogger) Debug(msg string, args ...interface{}) {
	if l.verbose {
		timestamp := time.Now().Format("15:04:05")
		fmt.Fprintf(l.output, "[%s] DEBUG: %s\n", timestamp, fmt.Sprintf(msg, args...))
	}
}
