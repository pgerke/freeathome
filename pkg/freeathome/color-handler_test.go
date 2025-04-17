package freeathome

import (
	"context"
	"io"
	"log/slog"
	"testing"
)

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

func TestColorHandler_Enabled(t *testing.T) {
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
