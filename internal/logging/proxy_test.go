// Package logging — Proxy та Bridge патерни для логування.
package logging

import (
	"bytes"
	"strings"
	"testing"
)

// TestLoggerProxy_LazyInit — перевіряє ліниву ініціалізацію реального логера.
func TestLoggerProxy_LazyInit(t *testing.T) {
	var buf bytes.Buffer
	proxy := NewLoggerProxy(&buf, "info")

	// До першого виклику Log — logCount має бути 0
	if proxy.GetLogCount() != 0 {
		t.Errorf("expected logCount 0 before any Log call, got %d", proxy.GetLogCount())
	}

	// realLogger має бути nil до першого виклику
	if proxy.realLogger != nil {
		t.Error("expected realLogger to be nil before first Log call")
	}

	proxy.Log("info", "test")

	// Після першого виклику — logCount має бути 1
	if proxy.GetLogCount() != 1 {
		t.Errorf("expected logCount 1 after one Log call, got %d", proxy.GetLogCount())
	}

	// realLogger має бути ініціалізований
	if proxy.realLogger == nil {
		t.Error("expected realLogger to be initialized after first Log call")
	}
}

// TestLoggerProxy_CountsLogs — перевіряє підрахунок записів у логі.
func TestLoggerProxy_CountsLogs(t *testing.T) {
	var buf bytes.Buffer
	proxy := NewLoggerProxy(&buf, "info")

	proxy.Log("info", "msg1")
	proxy.Log("warn", "msg2")
	proxy.Log("error", "msg3")

	if proxy.GetLogCount() != 3 {
		t.Errorf("expected logCount 3, got %d", proxy.GetLogCount())
	}
}

// TestLoggerProxy_LevelFiltering — перевіряє фільтрацію записів за рівнем.
func TestLoggerProxy_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	proxy := NewLoggerProxy(&buf, "warn")

	proxy.Log("info", "this should be skipped")
	proxy.Log("warn", "this should pass")
	proxy.Log("error", "this should also pass")

	if proxy.GetLogCount() != 2 {
		t.Errorf("expected logCount 2 (info skipped), got %d", proxy.GetLogCount())
	}

	if strings.Contains(buf.String(), "this should be skipped") {
		t.Error("expected info message to be filtered out, but it was written to buffer")
	}
}

// TestLoggerProxy_GetEntries — перевіряє отримання записів за рівнем.
func TestLoggerProxy_GetEntries(t *testing.T) {
	var buf bytes.Buffer
	proxy := NewLoggerProxy(&buf, "info")

	proxy.Log("info", "info message")
	proxy.Log("warn", "warn message")
	proxy.Log("error", "error message 1")
	proxy.Log("error", "error message 2")

	// Отримати тільки error записи
	errorEntries := proxy.GetEntries("error")
	if len(errorEntries) != 2 {
		t.Errorf("expected 2 error entries, got %d", len(errorEntries))
	}

	for _, e := range errorEntries {
		if e.Level != "error" {
			t.Errorf("expected all entries to have level 'error', got '%s'", e.Level)
		}
	}

	// Отримати всі записи
	allEntries := proxy.GetEntries("")
	if len(allEntries) != 4 {
		t.Errorf("expected 4 total entries, got %d", len(allEntries))
	}
}

// TestLoggerProxy_WritesToRealLogger — перевіряє що запис делегується реальному логеру.
func TestLoggerProxy_WritesToRealLogger(t *testing.T) {
	var buf bytes.Buffer
	proxy := NewLoggerProxy(&buf, "info")

	proxy.Log("info", "hello world")

	output := buf.String()
	if !strings.Contains(output, "hello world") {
		t.Errorf("expected buffer to contain 'hello world', got: %q", output)
	}
}
