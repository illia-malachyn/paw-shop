package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/illia-malachyn/paw-shop/internal/chat"
)

func TestHandleSend(t *testing.T) {
	h := NewChatHandler()

	body := `{"from":"customer1","to":"manager1","message":"Hello"}`
	req := httptest.NewRequest(http.MethodPost, "/api/chat/send", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleSend(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["status"].(string) != "sent" {
		t.Errorf("expected status 'sent', got '%s'", resp["status"])
	}
	if resp["from"].(string) != "customer1" {
		t.Errorf("expected from 'customer1', got '%s'", resp["from"])
	}
	if resp["to"].(string) != "manager1" {
		t.Errorf("expected to 'manager1', got '%s'", resp["to"])
	}
}

func TestHandleHistory(t *testing.T) {
	h := NewChatHandler()

	// Спочатку надсилаємо повідомлення
	sendBody := `{"from":"customer1","to":"manager1","message":"Hello"}`
	sendReq := httptest.NewRequest(http.MethodPost, "/api/chat/send", bytes.NewBufferString(sendBody))
	sendW := httptest.NewRecorder()
	h.HandleSend(sendW, sendReq)

	// Отримуємо історію для customer1
	req := httptest.NewRequest(http.MethodGet, "/api/chat/history?participant=customer1", nil)
	w := httptest.NewRecorder()

	h.HandleHistory(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var messages []chat.Message
	if err := json.NewDecoder(w.Body).Decode(&messages); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(messages) != 1 {
		t.Errorf("expected 1 message in history, got %d", len(messages))
	}

	if len(messages) == 1 {
		if messages[0].From != "customer1" {
			t.Errorf("expected from 'customer1', got '%s'", messages[0].From)
		}
		if messages[0].Content != "Hello" {
			t.Errorf("expected content 'Hello', got '%s'", messages[0].Content)
		}
	}
}

func TestHandleSendMissingFields(t *testing.T) {
	h := NewChatHandler()

	body := `{"from":"customer1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/chat/send", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleSend(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing to/message, got %d", w.Code)
	}
}

func TestHandleHistoryMissingParam(t *testing.T) {
	h := NewChatHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/chat/history", nil)
	w := httptest.NewRecorder()

	h.HandleHistory(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing participant, got %d", w.Code)
	}
}

func TestHandleSendMethodNotAllowed(t *testing.T) {
	h := NewChatHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/chat/send", nil)
	w := httptest.NewRecorder()

	h.HandleSend(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}
