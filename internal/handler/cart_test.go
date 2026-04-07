package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// addItemToCart — допоміжна функція для додавання товару до кошика в тестах.
func addItemToCart(t *testing.T, h *CartHandler, body string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/cart/add", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	h.HandleCart(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("helper addItemToCart: expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleCartAdd(t *testing.T) {
	h := NewCartHandler()

	body := `{"product_id":"p1","name":"Food","price":100,"quantity":2}`
	req := httptest.NewRequest(http.MethodPost, "/api/cart/add", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleCart(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["status"].(string) != "added" {
		t.Errorf("expected status='added', got '%v'", resp["status"])
	}
	items, ok := resp["cart"].([]interface{})
	if !ok {
		t.Fatalf("expected cart to be array, got %T", resp["cart"])
	}
	if len(items) != 1 {
		t.Errorf("expected 1 item in cart, got %d", len(items))
	}
}

func TestHandleCartAddInvalid(t *testing.T) {
	h := NewCartHandler()

	// Missing product_id
	body := `{"product_id":"","name":"Food","price":100,"quantity":2}`
	req := httptest.NewRequest(http.MethodPost, "/api/cart/add", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleCart(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing product_id, got %d", w.Code)
	}
}

func TestHandleCartRemove(t *testing.T) {
	h := NewCartHandler()

	addItemToCart(t, h, `{"product_id":"p1","name":"Food","price":100,"quantity":2}`)

	body := `{"product_id":"p1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/cart/remove", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleCart(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["status"].(string) != "removed" {
		t.Errorf("expected status='removed', got '%v'", resp["status"])
	}
	items, ok := resp["cart"].([]interface{})
	if !ok {
		t.Fatalf("expected cart to be array, got %T", resp["cart"])
	}
	if len(items) != 0 {
		t.Errorf("expected empty cart after remove, got %d items", len(items))
	}
}

func TestHandleCartRemoveNotFound(t *testing.T) {
	h := NewCartHandler()

	body := `{"product_id":"nonexistent"}`
	req := httptest.NewRequest(http.MethodPost, "/api/cart/remove", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleCart(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for nonexistent product_id, got %d", w.Code)
	}
}

func TestHandleCartGet(t *testing.T) {
	h := NewCartHandler()

	addItemToCart(t, h, `{"product_id":"p1","name":"Food","price":100,"quantity":1}`)
	addItemToCart(t, h, `{"product_id":"p2","name":"Treat","price":50,"quantity":3}`)

	req := httptest.NewRequest(http.MethodGet, "/api/cart", nil)
	w := httptest.NewRecorder()

	h.HandleCart(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var items []interface{}
	json.NewDecoder(w.Body).Decode(&items)

	if len(items) != 2 {
		t.Errorf("expected 2 items in cart, got %d", len(items))
	}
}

func TestHandleCartUndo(t *testing.T) {
	h := NewCartHandler()

	addItemToCart(t, h, `{"product_id":"p1","name":"Food","price":100,"quantity":2}`)

	req := httptest.NewRequest(http.MethodPost, "/api/cart/undo", nil)
	w := httptest.NewRecorder()

	h.HandleCart(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["status"].(string) != "undone" {
		t.Errorf("expected status='undone', got '%v'", resp["status"])
	}
	items, ok := resp["cart"].([]interface{})
	if !ok {
		t.Fatalf("expected cart to be array, got %T", resp["cart"])
	}
	if len(items) != 0 {
		t.Errorf("expected empty cart after undo, got %d items", len(items))
	}
}

func TestHandleCartUndoEmpty(t *testing.T) {
	h := NewCartHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/cart/undo", nil)
	w := httptest.NewRecorder()

	h.HandleCart(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty undo history, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["error"].(string) != "nothing to undo" {
		t.Errorf("expected error='nothing to undo', got '%v'", resp["error"])
	}
}

func TestHandleCartExportJSON(t *testing.T) {
	h := NewCartHandler()

	addItemToCart(t, h, `{"product_id":"p1","name":"Food","price":100,"quantity":2}`)

	req := httptest.NewRequest(http.MethodGet, "/api/cart/export?format=json", nil)
	w := httptest.NewRecorder()

	h.HandleCart(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("expected Content-Type application/json, got '%s'", ct)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Errorf("expected valid JSON body, got decode error: %v", err)
	}
}

func TestHandleCartExportText(t *testing.T) {
	h := NewCartHandler()

	addItemToCart(t, h, `{"product_id":"p1","name":"Food","price":100,"quantity":2}`)

	req := httptest.NewRequest(http.MethodGet, "/api/cart/export?format=text", nil)
	w := httptest.NewRecorder()

	h.HandleCart(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/plain") {
		t.Errorf("expected Content-Type text/plain, got '%s'", ct)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Receipt") {
		t.Errorf("expected body to contain 'Receipt', got: %s", body)
	}
	if !strings.Contains(body, "Total") {
		t.Errorf("expected body to contain 'Total', got: %s", body)
	}
}

func TestHandleCartExportInvalidFormat(t *testing.T) {
	h := NewCartHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/cart/export?format=xml", nil)
	w := httptest.NewRecorder()

	h.HandleCart(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid format, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["error"].(string) != "format must be json or text" {
		t.Errorf("expected error='format must be json or text', got '%v'", resp["error"])
	}
}

func TestHandleCartWrongMethod(t *testing.T) {
	h := NewCartHandler()

	// GET on /api/cart/add should return 405
	req := httptest.NewRequest(http.MethodGet, "/api/cart/add", nil)
	w := httptest.NewRecorder()

	h.HandleCart(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 for GET on /api/cart/add, got %d", w.Code)
	}
}
