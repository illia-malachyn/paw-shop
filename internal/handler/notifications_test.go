package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleNotify_Success(t *testing.T) {
	h := NewNotificationHandler()

	body := `{"user_id":"u1","message":"hello"}`
	req := httptest.NewRequest(http.MethodPost, "/api/notifications/send", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleNotify(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["status"] != "sent" {
		t.Errorf("expected status 'sent', got '%s'", resp["status"])
	}
}

func TestHandleNotify_MissingFields(t *testing.T) {
	h := NewNotificationHandler()

	body := `{"user_id":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/notifications/send", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleNotify(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleNotify_WrongMethod(t *testing.T) {
	h := NewNotificationHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/notifications/send", nil)
	w := httptest.NewRecorder()

	h.HandleNotify(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleLogs_ReturnsEntries(t *testing.T) {
	h := NewNotificationHandler()

	// Надсилаємо нотифікацію, щоб згенерувати запис у лозі
	notifyBody := `{"user_id":"u1","message":"hello"}`
	notifyReq := httptest.NewRequest(http.MethodPost, "/api/notifications/send", bytes.NewBufferString(notifyBody))
	notifyW := httptest.NewRecorder()
	h.HandleNotify(notifyW, notifyReq)

	// Отримуємо всі записи логу
	req := httptest.NewRequest(http.MethodGet, "/api/logs", nil)
	w := httptest.NewRecorder()

	h.HandleLogs(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var entries []map[string]string
	json.NewDecoder(w.Body).Decode(&entries)

	if len(entries) == 0 {
		t.Error("expected at least one log entry, got 0")
	}
}

func TestHandleLogs_FilterByLevel(t *testing.T) {
	h := NewNotificationHandler()

	// Надсилаємо кілька нотифікацій для генерації записів
	notifyBody := `{"user_id":"u2","message":"test"}`
	notifyReq := httptest.NewRequest(http.MethodPost, "/api/notifications/send", bytes.NewBufferString(notifyBody))
	h.HandleNotify(httptest.NewRecorder(), notifyReq)

	// Запит із фільтром level=info
	req := httptest.NewRequest(http.MethodGet, "/api/logs?level=info", nil)
	w := httptest.NewRecorder()

	h.HandleLogs(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var entries []map[string]string
	json.NewDecoder(w.Body).Decode(&entries)

	for _, e := range entries {
		if e["level"] != "info" {
			t.Errorf("expected all entries to have level 'info', got '%s'", e["level"])
		}
	}
}

func TestHandleLogStats_ReturnsCounts(t *testing.T) {
	h := NewNotificationHandler()

	// Надсилаємо нотифікацію для генерації записів
	notifyBody := `{"user_id":"u3","message":"stats test"}`
	notifyReq := httptest.NewRequest(http.MethodPost, "/api/notifications/send", bytes.NewBufferString(notifyBody))
	h.HandleNotify(httptest.NewRecorder(), notifyReq)

	req := httptest.NewRequest(http.MethodGet, "/api/logs/stats", nil)
	w := httptest.NewRecorder()

	h.HandleLogStats(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var stats map[string]int
	json.NewDecoder(w.Body).Decode(&stats)

	total, ok := stats["total"]
	if !ok {
		t.Fatal("expected 'total' key in stats response")
	}
	if total == 0 {
		t.Error("expected total > 0")
	}
}
