package notify

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

// failingWriter — io.Writer що завжди повертає помилку (для тестування error propagation).
type failingWriter struct{}

func (fw *failingWriter) Write(p []byte) (int, error) {
	return 0, errors.New("write failed")
}

// TestNotifyUser_SendsToBothChannels перевіряє що NotifyUser пише в обидва канали.
func TestNotifyUser_SendsToBothChannels(t *testing.T) {
	var consoleBuf, fileBuf bytes.Buffer
	facade := NewNotificationFacade(&consoleBuf, &fileBuf)

	err := facade.NotifyUser("user1", "test message")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	consoleOut := consoleBuf.String()
	fileOut := fileBuf.String()

	if !strings.Contains(consoleOut, "[CONSOLE]") {
		t.Errorf("console output missing [CONSOLE] prefix, got: %q", consoleOut)
	}
	if !strings.Contains(consoleOut, "user1") {
		t.Errorf("console output missing userID, got: %q", consoleOut)
	}
	if !strings.Contains(consoleOut, "test message") {
		t.Errorf("console output missing message, got: %q", consoleOut)
	}

	if !strings.Contains(fileOut, "[FILE]") {
		t.Errorf("file output missing [FILE] prefix, got: %q", fileOut)
	}
	if !strings.Contains(fileOut, "user1") {
		t.Errorf("file output missing userID, got: %q", fileOut)
	}
	if !strings.Contains(fileOut, "test message") {
		t.Errorf("file output missing message, got: %q", fileOut)
	}
}

// TestNotifyOrderStatusChanged_FormatsMessage перевіряє форматування повідомлення.
func TestNotifyOrderStatusChanged_FormatsMessage(t *testing.T) {
	var consoleBuf, fileBuf bytes.Buffer
	facade := NewNotificationFacade(&consoleBuf, &fileBuf)

	err := facade.NotifyOrderStatusChanged("order-1", "shipped")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	consoleOut := consoleBuf.String()

	if !strings.Contains(consoleOut, "order-1") {
		t.Errorf("output missing orderID, got: %q", consoleOut)
	}
	if !strings.Contains(consoleOut, "shipped") {
		t.Errorf("output missing status, got: %q", consoleOut)
	}
}

// TestFacade_ErrorPropagation перевіряє що facade повертає помилку якщо notifier завершується з помилкою.
func TestFacade_ErrorPropagation(t *testing.T) {
	fw := &failingWriter{}
	var goodBuf bytes.Buffer

	// ConsoleNotifier буде failing, FileNotifier — good
	facade := NewNotificationFacade(fw, &goodBuf)

	err := facade.NotifyUser("user1", "hello")
	if err == nil {
		t.Error("expected non-nil error when a notifier fails, got nil")
	}
}
