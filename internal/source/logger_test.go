package source

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewConsoleLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewConsoleLogger(buf, true)
	if logger == nil || logger.output != buf || !logger.verbose {
		t.Errorf("NewConsoleLogger did not initialize correctly")
	}
}

func TestConsoleLogger_Info(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewConsoleLogger(buf, false)
	logger.Info("test message %s", "arg")
	output := buf.String()
	if !strings.Contains(output, "INFO") || !strings.Contains(output, "test message arg") {
		t.Errorf("Info output incorrect: %s", output)
	}
}

func TestConsoleLogger_Error(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewConsoleLogger(buf, false)
	logger.Error("error message %d", 42)
	output := buf.String()
	if !strings.Contains(output, "ERROR") || !strings.Contains(output, "error message 42") {
		t.Errorf("Error output incorrect: %s", output)
	}
}

func TestConsoleLogger_Debug_verbose(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewConsoleLogger(buf, true)
	logger.Debug("debug message")
	output := buf.String()
	if !strings.Contains(output, "DEBUG") || !strings.Contains(output, "debug message") {
		t.Errorf("Debug output incorrect when verbose=true: %s", output)
	}
}

func TestConsoleLogger_Debug_notVerbose(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewConsoleLogger(buf, false)
	logger.Debug("debug message")
	output := buf.String()
	if output != "" {
		t.Errorf("Debug should not output when verbose=false, got: %s", output)
	}
}
