// Package logging — Bridge патерн для форматування та виводу логів.
// Bridge: Formatter (абстракція: що формат) × OutputWriter (реалізація: куди писати).
// Будь-який Formatter можна скомбінувати з будь-яким OutputWriter.
package logging

import (
	"fmt"
	"io"
	"os"
)

// OutputWriter — інтерфейс реалізатора (implementor) у патерні Bridge.
// Визначає куди виводити дані. Реалізації: ConsoleWriter, FileWriter.
type OutputWriter interface {
	Write(data string) error
}

// ConsoleWriter — виводить дані у writer (за замовчуванням os.Stdout). Патерн Bridge: реалізатор.
type ConsoleWriter struct {
	writer io.Writer
}

// NewConsoleWriter — конструктор ConsoleWriter. Якщо w == nil, використовує os.Stdout.
func NewConsoleWriter(w io.Writer) *ConsoleWriter {
	if w == nil {
		w = os.Stdout
	}
	return &ConsoleWriter{writer: w}
}

// Write — записує рядок у writer.
func (c *ConsoleWriter) Write(data string) error {
	_, err := fmt.Fprint(c.writer, data)
	return err
}

// FileWriter — записує дані у writer (файл або буфер). Патерн Bridge: реалізатор.
type FileWriter struct {
	writer io.Writer
}

// NewFileWriter — конструктор FileWriter.
func NewFileWriter(w io.Writer) *FileWriter {
	return &FileWriter{writer: w}
}

// Write — записує рядок у writer.
func (f *FileWriter) Write(data string) error {
	_, err := fmt.Fprint(f.writer, data)
	return err
}

// Formatter — базова абстракція у патерні Bridge. Містить посилання на реалізатор OutputWriter.
// Конкретні форматери (TextFormatter, JSONFormatter) вбудовують цю структуру.
type Formatter struct {
	writer OutputWriter
}

// TextFormatter — форматує повідомлення як plain text ([LEVEL] message). Патерн Bridge: абстракція.
type TextFormatter struct {
	Formatter
}

// NewTextFormatter — конструктор TextFormatter.
func NewTextFormatter(w OutputWriter) *TextFormatter {
	return &TextFormatter{Formatter{writer: w}}
}

// Format — форматує повідомлення як [LEVEL] message та передає у writer.
func (f *TextFormatter) Format(level, message string) error {
	formatted := fmt.Sprintf("[%s] %s\n", level, message)
	return f.writer.Write(formatted)
}

// JSONFormatter — форматує повідомлення як JSON. Патерн Bridge: абстракція.
type JSONFormatter struct {
	Formatter
}

// NewJSONFormatter — конструктор JSONFormatter.
func NewJSONFormatter(w OutputWriter) *JSONFormatter {
	return &JSONFormatter{Formatter{writer: w}}
}

// Format — форматує повідомлення як JSON та передає у writer.
func (f *JSONFormatter) Format(level, message string) error {
	formatted := fmt.Sprintf("{\"level\":%q,\"message\":%q}\n", level, message)
	return f.writer.Write(formatted)
}
