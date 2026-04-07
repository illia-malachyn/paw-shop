// Package logging — тести для Bridge патерну (Formatter × OutputWriter).
package logging

import (
	"bytes"
	"strings"
	"testing"
)

// TestBridge_AllCombinations — перевіряє всі 4 комбінації formatter + writer.
func TestBridge_AllCombinations(t *testing.T) {
	tests := []struct {
		name       string
		makeWriter func(buf *bytes.Buffer) OutputWriter
		makeFormat func(w OutputWriter) interface {
			Format(level, message string) error
		}
		wantContains string
	}{
		{
			name: "TextFormatter + ConsoleWriter",
			makeWriter: func(buf *bytes.Buffer) OutputWriter {
				return NewConsoleWriter(buf)
			},
			makeFormat: func(w OutputWriter) interface {
				Format(level, message string) error
			} {
				return NewTextFormatter(w)
			},
			wantContains: "[INFO] test message",
		},
		{
			name: "TextFormatter + FileWriter",
			makeWriter: func(buf *bytes.Buffer) OutputWriter {
				return NewFileWriter(buf)
			},
			makeFormat: func(w OutputWriter) interface {
				Format(level, message string) error
			} {
				return NewTextFormatter(w)
			},
			wantContains: "[INFO] test message",
		},
		{
			name: "JSONFormatter + ConsoleWriter",
			makeWriter: func(buf *bytes.Buffer) OutputWriter {
				return NewConsoleWriter(buf)
			},
			makeFormat: func(w OutputWriter) interface {
				Format(level, message string) error
			} {
				return NewJSONFormatter(w)
			},
			wantContains: `"level":"INFO"`,
		},
		{
			name: "JSONFormatter + FileWriter",
			makeWriter: func(buf *bytes.Buffer) OutputWriter {
				return NewFileWriter(buf)
			},
			makeFormat: func(w OutputWriter) interface {
				Format(level, message string) error
			} {
				return NewJSONFormatter(w)
			},
			wantContains: `"message":"test message"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			w := tc.makeWriter(&buf)
			f := tc.makeFormat(w)

			err := f.Format("INFO", "test message")
			if err != nil {
				t.Fatalf("Format returned error: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, tc.wantContains) {
				t.Errorf("expected output to contain %q, got: %q", tc.wantContains, output)
			}
		})
	}
}

// TestBridge_JSONFormat — перевіряє що JSONFormatter виводить валідний JSON (є { та }).
func TestBridge_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	w := NewConsoleWriter(&buf)
	f := NewJSONFormatter(w)

	err := f.Format("ERROR", "something failed")
	if err != nil {
		t.Fatalf("Format returned error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "{") || !strings.Contains(output, "}") {
		t.Errorf("expected JSON format with braces, got: %q", output)
	}
}

// TestBridge_TextFormat — перевіряє що TextFormatter виводить дужковий формат.
func TestBridge_TextFormat(t *testing.T) {
	var buf bytes.Buffer
	w := NewFileWriter(&buf)
	f := NewTextFormatter(w)

	err := f.Format("WARN", "something odd")
	if err != nil {
		t.Fatalf("Format returned error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[") || !strings.Contains(output, "]") {
		t.Errorf("expected text format with brackets, got: %q", output)
	}
	if !strings.Contains(output, "something odd") {
		t.Errorf("expected output to contain 'something odd', got: %q", output)
	}
}
