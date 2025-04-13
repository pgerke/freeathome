package freeathome

import "testing"

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
