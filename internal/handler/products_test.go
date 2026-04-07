package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/illia-malachyn/paw-shop/internal/models"
)

func TestHandleProducts_ReturnsAllProducts(t *testing.T) {
	h := NewProductHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/products", nil)
	w := httptest.NewRecorder()

	h.HandleProducts(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var products []models.ProductResponse
	if err := json.NewDecoder(w.Body).Decode(&products); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// 2 бренди * 3 категорії = 6 товарів
	if len(products) != 6 {
		t.Errorf("expected 6 products, got %d", len(products))
	}
}

func TestHandleProducts_HasCorrectBrands(t *testing.T) {
	h := NewProductHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/products", nil)
	w := httptest.NewRecorder()

	h.HandleProducts(w, req)

	var products []models.ProductResponse
	json.NewDecoder(w.Body).Decode(&products)

	brands := map[string]bool{}
	for _, p := range products {
		brands[p.Brand] = true
	}

	if !brands["Royal Canin"] {
		t.Error("expected Royal Canin brand in catalog")
	}
	if !brands["Acana"] {
		t.Error("expected Acana brand in catalog")
	}
}

func TestHandleProducts_HasAllCategories(t *testing.T) {
	h := NewProductHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/products", nil)
	w := httptest.NewRecorder()

	h.HandleProducts(w, req)

	var products []models.ProductResponse
	json.NewDecoder(w.Body).Decode(&products)

	categories := map[string]int{}
	for _, p := range products {
		categories[p.Category]++
	}

	if categories["dry"] != 2 {
		t.Errorf("expected 2 dry food products, got %d", categories["dry"])
	}
	if categories["wet"] != 2 {
		t.Errorf("expected 2 wet food products, got %d", categories["wet"])
	}
	if categories["treat"] != 2 {
		t.Errorf("expected 2 treat products, got %d", categories["treat"])
	}
}

func TestHandleProducts_MethodNotAllowed(t *testing.T) {
	h := NewProductHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/products", nil)
	w := httptest.NewRecorder()

	h.HandleProducts(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleProducts_ContentType(t *testing.T) {
	h := NewProductHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/products", nil)
	w := httptest.NewRecorder()

	h.HandleProducts(w, req)

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got '%s'", ct)
	}
}
