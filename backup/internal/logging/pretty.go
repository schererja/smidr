package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"
)

// PrettyHandler is a human-readable handler for development
type PrettyHandler struct {
	opts   slog.HandlerOptions
	output io.Writer
	attrs  []slog.Attr
	groups []string
	mu     sync.Mutex
}

// NewPrettyHandler creates a new pretty handler for clean debug output
func NewPrettyHandler(w io.Writer, opts *slog.HandlerOptions) *PrettyHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &PrettyHandler{
		opts:   *opts,
		output: w,
	}
}

func (h *PrettyHandler) Enabled(ctx context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

func (h *PrettyHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Format: TIME [LEVEL] message (key=val key=val)
	buf := &strings.Builder{}

	// Timestamp
	if !r.Time.IsZero() {
		buf.WriteString(r.Time.Format("15:04:05.000"))
		buf.WriteString(" ")
	}

	// Level with color
	levelStr := h.formatLevel(r.Level)
	buf.WriteString(levelStr)
	buf.WriteString(" ")

	// Message
	buf.WriteString(r.Message)

	// Attributes
	attrs := make([]string, 0)
	
	// Add handler's base attributes
	for _, attr := range h.attrs {
		attrs = append(attrs, h.formatAttr(attr))
	}

	// Add record attributes
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, h.formatAttr(a))
		return true
	})

	if len(attrs) > 0 {
		buf.WriteString(" ")
		buf.WriteString(strings.Join(attrs, " "))
	}

	buf.WriteString("\n")
	_, err := h.output.Write([]byte(buf.String()))
	return err
}

func (h *PrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)
	
	return &PrettyHandler{
		opts:   h.opts,
		output: h.output,
		attrs:  newAttrs,
		groups: h.groups,
	}
}

func (h *PrettyHandler) WithGroup(name string) slog.Handler {
	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name
	
	return &PrettyHandler{
		opts:   h.opts,
		output: h.output,
		attrs:  h.attrs,
		groups: newGroups,
	}
}

func (h *PrettyHandler) formatLevel(level slog.Level) string {
	switch level {
	case slog.LevelDebug:
		return "[DEBUG]"
	case slog.LevelInfo:
		return "[INFO] "
	case slog.LevelWarn:
		return "[WARN] "
	case slog.LevelError:
		return "[ERROR]"
	default:
		return fmt.Sprintf("[%s]", level.String())
	}
}

func (h *PrettyHandler) formatAttr(a slog.Attr) string {
	if a.Equal(slog.Attr{}) {
		return ""
	}

	key := a.Key
	if len(h.groups) > 0 {
		key = strings.Join(h.groups, ".") + "." + key
	}

	value := a.Value.String()
	
	// Special formatting for certain types
	switch a.Value.Kind() {
	case slog.KindTime:
		t := a.Value.Time()
		value = t.Format(time.RFC3339)
	case slog.KindDuration:
		value = a.Value.Duration().String()
	}

	return fmt.Sprintf("%s=%s", key, value)
}
