package freeathome

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"reflect"
	"strings"
	"testing"
	"time"
)

type DummyHandler struct {
	attrs  []slog.Attr
	groups []string
}

func (h *DummyHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}
func (h *DummyHandler) Handle(ctx context.Context, r slog.Record) error {
	return nil
}
func (h *DummyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.attrs = append(h.attrs, attrs...)
	return h
}
func (h *DummyHandler) WithGroup(name string) slog.Handler {
	h.groups = append(h.groups, name)
	return h
}

func TestNewColorHandler(t *testing.T) {
	tests := []struct {
		name       string
		out        io.Writer
		opts       *slog.HandlerOptions
		expectOpts bool
	}{
		{
			name:       "Nil options",
			out:        io.Discard,
			opts:       nil,
			expectOpts: true,
		},
		{
			name:       "Valid options",
			out:        io.Discard,
			opts:       &slog.HandlerOptions{Level: slog.LevelInfo},
			expectOpts: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := NewColorHandler(test.out, test.opts)

			if handler.out != test.out {
				t.Errorf("Expected output writer to be %v, got %v", test.out, handler.out)
			}

			if test.expectOpts && handler.opts == nil {
				t.Errorf("Expected options to be non-nil")
			}

			if test.opts != nil && handler.opts.Level != test.opts.Level {
				t.Errorf("Expected options level to be %v, got %v", test.opts.Level, handler.opts.Level)
			}

			if handler.base == nil {
				t.Errorf("Expected base handler to be initialized, got nil")
			}
		})
	}
}

func TestColorHandlerEnabled(t *testing.T) {
	tests := []struct {
		name     string
		level    slog.Level
		expected bool
	}{
		{
			name:     "Enabled for level Info",
			level:    slog.LevelInfo,
			expected: true,
		},
		{
			name:     "Disabled for level Debug",
			level:    slog.LevelDebug,
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := NewColorHandler(io.Discard, nil)

			result := handler.Enabled(context.Background(), test.level)
			if result != test.expected {
				t.Errorf("Enabled(%v) = %v; expected %v", test.level, result, test.expected)
			}
		})
	}
}

func TestColorHandlerHandle(t *testing.T) {
	tests := []struct {
		name       string
		record     slog.Record
		addSource  bool
		attributes []slog.Attr
	}{
		{
			name: "Debug log message",
			record: slog.Record{
				Time:    time.Now(),
				Level:   slog.LevelDebug,
				Message: "Debug message",
			},
			addSource:  true,
			attributes: nil,
		},
		{
			name: "Basic log message without source or attributes",
			record: slog.Record{
				Time:    time.Now(),
				Level:   slog.LevelInfo,
				Message: "Test message",
			},
			addSource:  false,
			attributes: nil,
		},
		{
			name: "Log message with source enabled",
			record: slog.Record{
				Time:    time.Now(),
				Level:   slog.LevelWarn,
				Message: "Warning message",
				PC:      1, // Simulate a program counter
			},
			addSource:  true,
			attributes: nil,
		},
		{
			name: "Log message with attributes",
			record: slog.Record{
				Time:    time.Now(),
				Level:   slog.LevelError,
				Message: "Error occurred",
			},
			addSource: false,
			attributes: []slog.Attr{
				{Key: "key1", Value: slog.StringValue("value1")},
				{Key: "key2", Value: slog.StringValue("value2")},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output strings.Builder
			handler := &ColorHandler{
				out: io.Writer(&output),
				opts: &slog.HandlerOptions{
					AddSource: test.addSource,
				},
				base: slog.NewTextHandler(io.Discard, nil),
			}

			// Add attributes to the record if provided
			if test.attributes != nil {
				for _, attr := range test.attributes {
					test.record.AddAttrs(attr)
				}
			}

			err := handler.Handle(context.Background(), test.record)
			if err != nil {
				t.Fatalf("Handle returned an error: %v", err)
			}

			// Validate the output contains expected components
			if !strings.Contains(output.String(), "time=") {
				t.Errorf("Expected output to contain timestamp, got: %s", output.String())
			}
			if !strings.Contains(output.String(), fmt.Sprintf("level=%s", test.record.Level.String())) {
				t.Errorf("Expected output to contain level, got: %s", output.String())
			}
			if !strings.Contains(output.String(), fmt.Sprintf("msg=%s", logfmtEscape(test.record.Message))) {
				t.Errorf("Expected output to contain message, got: %s", output.String())
			}
			if test.addSource && test.record.PC != 0 && !strings.Contains(output.String(), "source=") {
				t.Errorf("Expected output to contain source, got: %s", output.String())
			}
			for _, attr := range test.attributes {
				if !strings.Contains(output.String(), fmt.Sprintf("%s=%s", attr.Key, logfmtEscape(fmt.Sprint(attr.Value)))) {
					t.Errorf("Expected output to contain attribute %s=%s, got: %s", attr.Key, attr.Value, output.String())
				}
			}
		})
	}
}

func TestColorHandlerWithAttrs(t *testing.T) {
	tests := []struct {
		name  string
		attrs []slog.Attr
	}{
		{
			name:  "No attributes",
			attrs: nil,
		},
		{
			name: "Single attribute",
			attrs: []slog.Attr{
				{Key: "key1", Value: slog.StringValue("value1")},
			},
		},
		{
			name: "Multiple attributes",
			attrs: []slog.Attr{
				{Key: "key1", Value: slog.StringValue("value1")},
				{Key: "key2", Value: slog.StringValue("value2")},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			base := &DummyHandler{}
			handler := &ColorHandler{
				out:  io.Discard,
				opts: nil,
				base: base,
			}
			newHandler := handler.WithAttrs(test.attrs)

			if newHandler == nil {
				t.Fatalf("WithAttrs returned nil")
			}

			colorHandler, ok := newHandler.(*ColorHandler)
			if !ok {
				t.Fatalf("Expected handler to be of type *ColorHandler, got %T", newHandler)
			}

			if colorHandler.out != handler.out {
				t.Errorf("Expected output writer to be %v, got %v", handler.out, colorHandler.out)
			}

			if colorHandler.opts != handler.opts {
				t.Errorf("Expected options to be %v, got %v", handler.opts, colorHandler.opts)
			}

			if !reflect.DeepEqual(base.attrs, test.attrs) {
				t.Errorf("Expected attributes to be %v, got %v", test.attrs, base.attrs)
			}
		})
	}
}

func TestColorHandlerWithGroup(t *testing.T) {
	tests := []struct {
		name      string
		groupName string
	}{
		{
			name:      "Empty group name",
			groupName: "",
		},
		{
			name:      "Valid group name",
			groupName: "testGroup",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			base := &DummyHandler{}
			handler := &ColorHandler{
				out:  io.Discard,
				opts: nil,
				base: base,
			}
			newHandler := handler.WithGroup(test.groupName)

			if newHandler == nil {
				t.Fatalf("WithGroup returned nil")
			}

			colorHandler, ok := newHandler.(*ColorHandler)
			if !ok {
				t.Fatalf("Expected handler to be of type *ColorHandler, got %T", newHandler)
			}

			if colorHandler.out != handler.out {
				t.Errorf("Expected output writer to be %v, got %v", handler.out, colorHandler.out)
			}

			if colorHandler.opts != handler.opts {
				t.Errorf("Expected options to be %v, got %v", handler.opts, colorHandler.opts)
			}

			if len(base.groups) != 1 {
				t.Errorf("Expected group legth to be 1, got %v", len(base.groups))
			}

			if base.groups[0] != test.groupName {
				t.Errorf("Expected group name to be %v, got %v", test.groupName, base.groups[0])
			}
		})
	}
}

func TestLogFmtEscape(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello!", "Hello!"},
		{"Hello, World!", "\"Hello, World!\""},
	}

	for _, test := range tests {
		result := logfmtEscape(test.input)
		if result != test.expected {
			t.Errorf("logfmtEscape(%q) = %q; expected %q", test.input, result, test.expected)
		}
	}
}
