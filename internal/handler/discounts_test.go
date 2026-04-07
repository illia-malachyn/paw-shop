package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleApply_Percent(t *testing.T) {
	h := NewDiscountHandler()

	body := `{"product_id":"rc-dry-01","discount_type":"percent","value":10}`
	req := httptest.NewRequest(http.MethodPost, "/api/discounts/apply", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleApply(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["new_price"].(float64) != 1305 {
		t.Errorf("expected 1305 (10%% off 1450), got %.2f", resp["new_price"].(float64))
	}
	if resp["discount"].(string) != "percent" {
		t.Errorf("expected discount 'percent', got '%s'", resp["discount"].(string))
	}
}

func TestHandleApply_Fixed(t *testing.T) {
	h := NewDiscountHandler()

	body := `{"product_id":"rc-wet-01","discount_type":"fixed","value":100}`
	req := httptest.NewRequest(http.MethodPost, "/api/discounts/apply", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleApply(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["new_price"].(float64) != 420 {
		t.Errorf("expected 420 (520-100), got %.2f", resp["new_price"].(float64))
	}
}

func TestHandleApply_UnknownType(t *testing.T) {
	h := NewDiscountHandler()

	body := `{"product_id":"rc-dry-01","discount_type":"bogus","value":10}`
	req := httptest.NewRequest(http.MethodPost, "/api/discounts/apply", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleApply(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for unknown discount type, got %d", w.Code)
	}
}

func TestHandleApply_InvalidJSON(t *testing.T) {
	h := NewDiscountHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/discounts/apply", bytes.NewBufferString("{bad"))
	w := httptest.NewRecorder()

	h.HandleApply(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleApply_MethodNotAllowed(t *testing.T) {
	h := NewDiscountHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/discounts/apply", nil)
	w := httptest.NewRecorder()

	h.HandleApply(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleUndo(t *testing.T) {
	h := NewDiscountHandler()

	// Спочатку застосуємо знижку
	applyBody := `{"product_id":"rc-dry-01","discount_type":"fixed","value":200}`
	applyReq := httptest.NewRequest(http.MethodPost, "/api/discounts/apply", bytes.NewBufferString(applyBody))
	applyW := httptest.NewRecorder()
	h.HandleApply(applyW, applyReq)

	// Тепер undo
	undoReq := httptest.NewRequest(http.MethodPost, "/api/discounts/undo", nil)
	undoW := httptest.NewRecorder()
	h.HandleUndo(undoW, undoReq)

	if undoW.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", undoW.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(undoW.Body).Decode(&resp)

	if resp["restored_price"].(float64) != 1450 {
		t.Errorf("expected restored price 1450, got %.2f", resp["restored_price"].(float64))
	}
}

func TestHandleUndo_EmptyHistory(t *testing.T) {
	h := NewDiscountHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/discounts/undo", nil)
	w := httptest.NewRecorder()

	h.HandleUndo(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty history, got %d", w.Code)
	}
}

func TestHandleSubscribe(t *testing.T) {
	h := NewDiscountHandler()

	body := `{"email":"test@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/products/rc-dry-01/subscribe", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleSubscribe(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["subscribed"].(bool) != true {
		t.Error("expected subscribed=true")
	}
	if resp["product_id"].(string) != "rc-dry-01" {
		t.Errorf("expected product_id 'rc-dry-01', got '%s'", resp["product_id"].(string))
	}
}

func TestHandleSubscribe_MissingEmail(t *testing.T) {
	h := NewDiscountHandler()

	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/api/products/rc-dry-01/subscribe", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleSubscribe(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing email, got %d", w.Code)
	}
}
