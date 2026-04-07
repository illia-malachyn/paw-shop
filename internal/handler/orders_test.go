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

// --- HandleStatus tests ---

func TestHandleStatus_NextAdvancesToConfirmed(t *testing.T) {
	h := NewOrderHandler()

	body := `{"action":"next"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/orders/order-1/status", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleOrders(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["status"].(string) != "confirmed" {
		t.Errorf("expected status='confirmed', got '%s'", resp["status"])
	}
	if resp["id"].(string) != "order-1" {
		t.Errorf("expected id='order-1', got '%s'", resp["id"])
	}
}

func TestHandleStatus_NextTwiceAdvancesToShipped(t *testing.T) {
	h := NewOrderHandler()

	// First next: new -> confirmed
	body := `{"action":"next"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/orders/order-1/status", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	h.HandleOrders(w, req)

	// Second next: confirmed -> shipped
	body = `{"action":"next"}`
	req = httptest.NewRequest(http.MethodPatch, "/api/orders/order-1/status", bytes.NewBufferString(body))
	w = httptest.NewRecorder()
	h.HandleOrders(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["status"].(string) != "shipped" {
		t.Errorf("expected status='shipped', got '%s'", resp["status"])
	}
}

func TestHandleStatus_CancelReturnsStatusCancelled(t *testing.T) {
	h := NewOrderHandler()

	body := `{"action":"cancel"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/orders/order-1/status", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleOrders(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["status"].(string) != "cancelled" {
		t.Errorf("expected status='cancelled', got '%s'", resp["status"])
	}
}

func TestHandleStatus_NotFoundReturns404(t *testing.T) {
	h := NewOrderHandler()

	body := `{"action":"next"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/orders/nonexistent/status", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleOrders(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for nonexistent order, got %d", w.Code)
	}
}

func TestHandleStatus_InvalidActionReturns400(t *testing.T) {
	h := NewOrderHandler()

	body := `{"action":"ship"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/orders/order-1/status", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleOrders(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid action, got %d", w.Code)
	}
}

func TestHandleStatus_WrongMethodReturns405(t *testing.T) {
	h := NewOrderHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/orders/order-1/status", nil)
	w := httptest.NewRecorder()

	h.HandleOrders(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 for GET on status endpoint, got %d", w.Code)
	}
}

// --- HandleListOrders tests ---

func TestHandleListOrders_ReturnsAll3Orders(t *testing.T) {
	h := NewOrderHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/orders", nil)
	w := httptest.NewRecorder()

	h.HandleOrders(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var orders []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&orders)

	if len(orders) != 3 {
		t.Errorf("expected 3 orders, got %d", len(orders))
	}
}

func TestHandleListOrders_FilteredByStatusNew(t *testing.T) {
	h := NewOrderHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/orders?status=new", nil)
	w := httptest.NewRecorder()

	h.HandleOrders(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var orders []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&orders)

	if len(orders) != 3 {
		t.Errorf("expected 3 orders with status=new, got %d", len(orders))
	}
	for _, o := range orders {
		if o["status"].(string) != "new" {
			t.Errorf("expected all orders to have status 'new', got '%s'", o["status"])
		}
	}
}

func TestHandleListOrders_FilteredByNonexistentStatusReturnsEmpty(t *testing.T) {
	h := NewOrderHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/orders?status=nonexistent", nil)
	w := httptest.NewRecorder()

	h.HandleOrders(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var orders []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&orders)

	if len(orders) != 0 {
		t.Errorf("expected 0 orders for nonexistent status, got %d", len(orders))
	}
}

// --- HandleCreateOrder tests ---

func TestHandleCreateOrder_ValidDataReturns201(t *testing.T) {
	h := NewOrderHandler()

	body := `{"items":["kibble"],"address":"123 Dog St","amount":50.0}`
	req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleOrders(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["id"] == nil {
		t.Error("expected response to contain 'id'")
	}
	if resp["status"].(string) != "new" {
		t.Errorf("expected status='new', got '%s'", resp["status"])
	}
	items, ok := resp["items"].([]interface{})
	if !ok || len(items) == 0 {
		t.Errorf("expected non-empty items in response, got %v", resp["items"])
	}
}

func TestHandleCreateOrder_OutOfStockItemReturns400(t *testing.T) {
	h := NewOrderHandler()

	body := `{"items":["out-of-stock-item"],"address":"123 Dog St","amount":50.0}`
	req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleOrders(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for out-of-stock item, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "stock") {
		t.Errorf("expected stock error in response, got: %s", w.Body.String())
	}
}

func TestHandleCreateOrder_EmptyAddressReturns400(t *testing.T) {
	h := NewOrderHandler()

	body := `{"items":["kibble"],"address":"","amount":50.0}`
	req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleOrders(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty address, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "address") {
		t.Errorf("expected address error in response, got: %s", w.Body.String())
	}
}

func TestHandleCreateOrder_ZeroAmountReturns400(t *testing.T) {
	h := NewOrderHandler()

	body := `{"items":["kibble"],"address":"123 Dog St","amount":0}`
	req := httptest.NewRequest(http.MethodPost, "/api/orders", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleOrders(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for zero amount, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "payment") {
		t.Errorf("expected payment error in response, got: %s", w.Body.String())
	}
}
