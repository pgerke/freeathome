package freeathome

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

// ColorHandler wraps a slog.Handler and colors the level based on severity
type ColorHandler struct {
	out  io.Writer
	opts *slog.HandlerOptions
	base slog.Handler
}

// NewColorHandler creates a colorized slog handler
func NewColorHandler(out io.Writer, opts *slog.HandlerOptions) *ColorHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{} // prevent fallback
	}

	return &ColorHandler{
		out:  out,
		opts: opts,
		base: slog.NewTextHandler(out, opts),
	}
}

// Enabled checks if the handler is enabled for the given level
func (h *ColorHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.base.Enabled(ctx, level)
}

// Handle formats the log message with colors and prints it to the console
func (h *ColorHandler) Handle(ctx context.Context, r slog.Record) error {
	// Timestamp in gray
	timestampColor := color.New(color.FgWhite).SprintFunc()
	timestamp := timestampColor(fmt.Sprintf("time=%s", r.Time.Format(time.RFC3339)))

	// Level in color
	levelColor := levelColor(r.Level)
	levelText := levelColor(fmt.Sprintf("level=%s", r.Level.String()))

	// Message in cyan
	msgColor := color.New(color.FgHiWhite).SprintFunc()
	message := msgColor(fmt.Sprintf("msg=%s", logfmtEscape(fmt.Sprint(r.Message))))

	// Source in magenta
	sourceColor := color.New(color.FgMagenta).SprintFunc()
	sourceText := ""
	if h.opts.AddSource && r.PC != 0 {
		frame, _ := runtime.CallersFrames([]uintptr{r.PC}).Next()
		sourceText = sourceColor(fmt.Sprintf(" source=%s:%d", frame.File, frame.Line))
	}

	// Attributes in green and yellow
	attrText := ""
	r.Attrs(func(a slog.Attr) bool {
		keyColor := color.New(color.FgCyan).SprintFunc()
		valColor := color.New(color.FgHiGreen).SprintFunc()
		attrText += fmt.Sprintf(" %s=%s", keyColor(a.Key), valColor(logfmtEscape(fmt.Sprint(a.Value))))
		return true
	})

	// Print the formatted log message
	_, err := fmt.Fprintf(h.out, "%s %s%s %s%s\n", timestamp, levelText, sourceText, message, attrText)
	return err
}

// WithAttrs adds attributes to the handler
func (h *ColorHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ColorHandler{
		out:  h.out,
		opts: h.opts,
		base: h.base.WithAttrs(attrs),
	}
}

// WithGroup adds a group to the handler
func (h *ColorHandler) WithGroup(name string) slog.Handler {
	return &ColorHandler{
		out:  h.out,
		opts: h.opts,
		base: h.base.WithGroup(name),
	}
}

// levelColor returns a color function based on the log level
func levelColor(level slog.Level) func(a ...any) string {
	switch {
	case level >= slog.LevelError:
		return color.New(color.FgRed, color.Bold).SprintFunc()
	case level == slog.LevelWarn:
		return color.New(color.FgYellow, color.Bold).SprintFunc()
	case level == slog.LevelInfo:
		return color.New(color.FgGreen).SprintFunc()
	default:
		return color.New(color.FgBlue).SprintFunc()
	}
}

func logfmtEscape(val string) string {
	if strings.ContainsAny(val, " \t\n\"") {
		val = strconv.Quote(val)
	}
	return val
}
