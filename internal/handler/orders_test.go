package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleBatch_Confirm(t *testing.T) {
	h := NewOrderHandler()

	body := `{"order_ids":["order-1","order-2"],"action":"confirm"}`
	req := httptest.NewRequest(http.MethodPost, "/api/orders/batch", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleBatch(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if int(resp["processed"].(float64)) != 2 {
		t.Errorf("expected processed=2, got %v", resp["processed"])
	}
	if resp["action"].(string) != "confirm" {
		t.Errorf("expected action='confirm', got '%v'", resp["action"])
	}

	orders := h.GetOrders()
	if orders["order-1"].Status != "confirmed" {
		t.Errorf("expected order-1 status 'confirmed', got '%s'", orders["order-1"].Status)
	}
	if orders["order-2"].Status != "confirmed" {
		t.Errorf("expected order-2 status 'confirmed', got '%s'", orders["order-2"].Status)
	}
}

func TestHandleBatch_Reject(t *testing.T) {
	h := NewOrderHandler()

	body := `{"order_ids":["order-1"],"action":"reject"}`
	req := httptest.NewRequest(http.MethodPost, "/api/orders/batch", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleBatch(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if int(resp["processed"].(float64)) != 1 {
		t.Errorf("expected processed=1, got %v", resp["processed"])
	}
}

func TestHandleBatch_InvalidAction(t *testing.T) {
	h := NewOrderHandler()

	body := `{"order_ids":["order-1"],"action":"invalid"}`
	req := httptest.NewRequest(http.MethodPost, "/api/orders/batch", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleBatch(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid action, got %d", w.Code)
	}
}

func TestHandleBatch_OrderNotFound(t *testing.T) {
	h := NewOrderHandler()

	body := `{"order_ids":["nonexistent"],"action":"confirm"}`
	req := httptest.NewRequest(http.MethodPost, "/api/orders/batch", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleBatch(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for nonexistent order, got %d", w.Code)
	}
}

func TestHandleBatch_WrongMethod(t *testing.T) {
	h := NewOrderHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/orders/batch", nil)
	w := httptest.NewRecorder()

	h.HandleBatch(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleReport_Daily(t *testing.T) {
	h := NewOrderHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/reports/daily", nil)
	w := httptest.NewRecorder()

	h.HandleReport(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["report_type"].(string) != "daily" {
		t.Errorf("expected report_type='daily', got '%v'", resp["report_type"])
	}
	content, ok := resp["content"].(string)
	if !ok || len(content) == 0 {
		t.Errorf("expected non-empty content, got %v", resp["content"])
	}
	if !strings.Contains(content, "Daily Report") {
		t.Errorf("expected content to contain 'Daily Report', got: %s", content)
	}
}

func TestHandleReport_Weekly(t *testing.T) {
	h := NewOrderHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/reports/weekly", nil)
	w := httptest.NewRecorder()

	h.HandleReport(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["report_type"].(string) != "weekly" {
		t.Errorf("expected report_type='weekly', got '%v'", resp["report_type"])
	}
	content, ok := resp["content"].(string)
	if !ok || len(content) == 0 {
		t.Errorf("expected non-empty content, got %v", resp["content"])
	}
	if !strings.Contains(content, "Weekly Report") {
		t.Errorf("expected content to contain 'Weekly Report', got: %s", content)
	}
}

func TestHandleReport_Unknown(t *testing.T) {
	h := NewOrderHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/reports/unknown", nil)
	w := httptest.NewRecorder()

	h.HandleReport(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for unknown report type, got %d", w.Code)
	}
}

func TestHandleReport_WrongMethod(t *testing.T) {
	h := NewOrderHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/reports/daily", nil)
	w := httptest.NewRecorder()

	h.HandleReport(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}
