package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/illia-malachyn/paw-shop/internal/models"
)

func TestHandleSearchBrand(t *testing.T) {
	h := NewSearchHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/products/search?q=brand:Royal", nil)
	w := httptest.NewRecorder()

	h.HandleSearch(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var products []models.ProductResponse
	if err := json.NewDecoder(w.Body).Decode(&products); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Повинні бути тільки товари Royal Canin (3 штуки)
	if len(products) != 3 {
		t.Errorf("expected 3 Royal Canin products, got %d", len(products))
	}

	for _, p := range products {
		if p.Brand != "Royal Canin" {
			t.Errorf("expected brand Royal Canin, got %s", p.Brand)
		}
	}
}

func TestHandleSearchPrice(t *testing.T) {
	h := NewSearchHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/products/search?q=price:%3C500", nil)
	w := httptest.NewRecorder()

	h.HandleSearch(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var products []models.ProductResponse
	if err := json.NewDecoder(w.Body).Decode(&products); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Royal Canin Dental Sticks (280) та Acana Crunchy Biscuits (350) < 500
	if len(products) != 2 {
		t.Errorf("expected 2 products under 500, got %d", len(products))
	}

	for _, p := range products {
		if p.Price >= 500 {
			t.Errorf("expected price < 500, got %.2f for %s", p.Price, p.Name)
		}
	}
}

func TestHandleSearchCategory(t *testing.T) {
	h := NewSearchHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/products/search?q=category:dry", nil)
	w := httptest.NewRecorder()

	h.HandleSearch(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var products []models.ProductResponse
	if err := json.NewDecoder(w.Body).Decode(&products); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Royal Canin Maxi Adult (dry) та Acana Wild Prairie (dry) = 2 товари
	if len(products) != 2 {
		t.Errorf("expected 2 dry food products, got %d", len(products))
	}

	for _, p := range products {
		if p.Category != "dry" {
			t.Errorf("expected category dry, got %s", p.Category)
		}
	}
}

func TestHandleSearchCombined(t *testing.T) {
	h := NewSearchHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/products/search?q=brand:Royal+AND+price:%3C500", nil)
	w := httptest.NewRecorder()

	h.HandleSearch(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var products []models.ProductResponse
	if err := json.NewDecoder(w.Body).Decode(&products); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Тільки Royal Canin Dental Sticks (280) відповідає обом умовам
	if len(products) != 1 {
		t.Errorf("expected 1 product matching brand:Royal AND price:<500, got %d", len(products))
	}

	if len(products) == 1 {
		if products[0].Brand != "Royal Canin" {
			t.Errorf("expected Royal Canin brand, got %s", products[0].Brand)
		}
		if products[0].Price >= 500 {
			t.Errorf("expected price < 500, got %.2f", products[0].Price)
		}
	}
}

func TestHandleSearchMissingQuery(t *testing.T) {
	h := NewSearchHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/products/search", nil)
	w := httptest.NewRecorder()

	h.HandleSearch(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing query, got %d", w.Code)
	}
}

func TestHandleSearchInvalidQuery(t *testing.T) {
	h := NewSearchHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/products/search?q=invalid", nil)
	w := httptest.NewRecorder()

	h.HandleSearch(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid query format, got %d", w.Code)
	}
}

func TestHandleSearchMethodNotAllowed(t *testing.T) {
	h := NewSearchHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/products/search", nil)
	w := httptest.NewRecorder()

	h.HandleSearch(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}
