// Package logging — Proxy та Bridge патерни для логування.
// Proxy: LoggerProxy контролює доступ до FileLogger (лінива ініціалізація, підрахунок, фільтрація).
// Bridge: Formatter абстракція × OutputWriter реалізація (bridge.go).
package logging

import (
	"fmt"
	"io"
	"time"
)

// LogEntry — запис у журналі логування.
type LogEntry struct {
	Level     string `json:"level"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// Logger — інтерфейс логера. Proxy pattern: дозволяє підмінити реальний логер проксі.
type Logger interface {
	Log(level, message string)
}

// FileLogger — реальний логер, що записує у io.Writer. Патерн Proxy: реальний об'єкт.
type FileLogger struct {
	writer io.Writer
}

// NewFileLogger — конструктор FileLogger.
func NewFileLogger(w io.Writer) *FileLogger {
	return &FileLogger{writer: w}
}

// Log — записує рядок у форматі [LEVEL] message у writer.
func (l *FileLogger) Log(level, message string) {
	fmt.Fprintf(l.writer, "[%s] %s\n", level, message)
}

// levelValue — допоміжна функція для порівняння рівнів логування.
// info=0, warn=1, error=2. Невідомий рівень повертає -1.
func levelValue(level string) int {
	switch level {
	case "info":
		return 0
	case "warn":
		return 1
	case "error":
		return 2
	default:
		return -1
	}
}

// LoggerProxy — проксі для Logger. Патерн Proxy: додає лінню ініціалізацію, підрахунок викликів та фільтрацію за рівнем.
type LoggerProxy struct {
	realLogger Logger
	writer     io.Writer
	minLevel   string
	entries    []LogEntry
	logCount   int
}

// NewLoggerProxy — конструктор LoggerProxy. realLogger залишається nil до першого виклику Log.
func NewLoggerProxy(w io.Writer, minLevel string) *LoggerProxy {
	return &LoggerProxy{
		writer:   w,
		minLevel: minLevel,
		entries:  []LogEntry{},
	}
}

// Log — делегує запис реальному логеру (lazy init), рахує записи та фільтрує за рівнем.
func (p *LoggerProxy) Log(level, message string) {
	// Фільтрація за мінімальним рівнем
	if levelValue(level) < levelValue(p.minLevel) {
		return
	}

	// Лінна ініціалізація реального логера
	if p.realLogger == nil {
		p.realLogger = NewFileLogger(p.writer)
	}

	// Делегування реальному логеру
	p.realLogger.Log(level, message)

	// Збереження запису в пам'яті
	p.entries = append(p.entries, LogEntry{
		Level:     level,
		Message:   message,
		Timestamp: time.Now().Format(time.RFC3339),
	})
	p.logCount++
}

// GetLogCount — повертає кількість записів у лозі.
func (p *LoggerProxy) GetLogCount() int {
	return p.logCount
}

// GetEntries — повертає записи з логу. Якщо level порожній — всі записи, інакше — тільки з вказаним рівнем.
func (p *LoggerProxy) GetEntries(level string) []LogEntry {
	if level == "" {
		return p.entries
	}

	filtered := []LogEntry{}
	for _, e := range p.entries {
		if e.Level == level {
			filtered = append(filtered, e)
		}
	}
	return filtered
}
