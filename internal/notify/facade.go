// Package notify реалізує патерн Facade для відправки нотифікацій через кілька каналів.
package notify

import (
	"errors"
	"fmt"
	"io"
	"os"
)

// Notifier — інтерфейс підсистеми нотифікацій.
type Notifier interface {
	Notify(userID, message string) error
}

// ConsoleNotifier — підсистема: виводить нотифікацію в консоль.
type ConsoleNotifier struct {
	writer io.Writer
}

// NewConsoleNotifier — конструктор ConsoleNotifier з виводом у os.Stdout за замовчуванням.
func NewConsoleNotifier() *ConsoleNotifier {
	return &ConsoleNotifier{writer: os.Stdout}
}

// Notify записує повідомлення у вигляді [CONSOLE] User userID: message.
func (n *ConsoleNotifier) Notify(userID, message string) error {
	_, err := fmt.Fprintf(n.writer, "[CONSOLE] User %s: %s\n", userID, message)
	return err
}

// FileNotifier — підсистема: записує нотифікацію у writer (у production — файл).
type FileNotifier struct {
	writer io.Writer
}

// NewFileNotifier — конструктор FileNotifier з переданим io.Writer.
func NewFileNotifier(w io.Writer) *FileNotifier {
	return &FileNotifier{writer: w}
}

// Notify записує повідомлення у вигляді [FILE] User userID: message.
func (n *FileNotifier) Notify(userID, message string) error {
	_, err := fmt.Fprintf(n.writer, "[FILE] User %s: %s\n", userID, message)
	return err
}

// NotificationFacade — фасад, який приховує складність роботи з кількома каналами нотифікацій.
type NotificationFacade struct {
	notifiers []Notifier
}

// NewNotificationFacade — конструктор NotificationFacade з ConsoleNotifier та FileNotifier.
func NewNotificationFacade(console io.Writer, file io.Writer) *NotificationFacade {
	return &NotificationFacade{
		notifiers: []Notifier{
			&ConsoleNotifier{writer: console},
			&FileNotifier{writer: file},
		},
	}
}

// NotifyUser відправляє повідомлення через усі канали нотифікацій, агрегуючи помилки.
func (f *NotificationFacade) NotifyUser(userID, message string) error {
	var errs []error
	for _, n := range f.notifiers {
		if err := n.Notify(userID, message); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// NotifyOrderStatusChanged форматує повідомлення про зміну статусу замовлення і надсилає через всі канали.
func (f *NotificationFacade) NotifyOrderStatusChanged(orderID, status string) error {
	message := fmt.Sprintf("Order %s status changed to %s", orderID, status)
	return f.NotifyUser("system", message)
}
