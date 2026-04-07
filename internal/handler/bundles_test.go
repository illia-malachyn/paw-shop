package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/illia-malachyn/paw-shop/internal/bundle"
)

func TestHandleTemplates(t *testing.T) {
	h := NewBundleHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/bundles/templates", nil)
	w := httptest.NewRecorder()

	h.HandleTemplates(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var templates map[string]string
	if err := json.NewDecoder(w.Body).Decode(&templates); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if _, ok := templates["puppy"]; !ok {
		t.Error("expected 'puppy' template")
	}
	if _, ok := templates["large_breed"]; !ok {
		t.Error("expected 'large_breed' template")
	}
	if _, ok := templates["senior"]; !ok {
		t.Error("expected 'senior' template")
	}
}

func TestHandleTemplates_MethodNotAllowed(t *testing.T) {
	h := NewBundleHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/bundles/templates", nil)
	w := httptest.NewRecorder()

	h.HandleTemplates(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleClone_Success(t *testing.T) {
	h := NewBundleHandler()

	body := `{"template":"puppy","name":"Мій набір для цуценяти"}`
	req := httptest.NewRequest(http.MethodPost, "/api/bundles/clone", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleClone(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	var b bundle.Bundle
	json.NewDecoder(w.Body).Decode(&b)

	if b.Name != "Мій набір для цуценяти" {
		t.Errorf("expected custom name, got '%s'", b.Name)
	}
	if b.DogSize != "small" {
		t.Errorf("expected dog_size 'small' from puppy template, got '%s'", b.DogSize)
	}
}

func TestHandleClone_OverrideExtras(t *testing.T) {
	h := NewBundleHandler()

	body := `{"template":"puppy","extras":["bowl"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/bundles/clone", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleClone(w, req)

	var b bundle.Bundle
	json.NewDecoder(w.Body).Decode(&b)

	if len(b.Extras) != 1 || b.Extras[0] != "bowl" {
		t.Errorf("expected extras [bowl], got %v", b.Extras)
	}
}

func TestHandleClone_TemplateNotFound(t *testing.T) {
	h := NewBundleHandler()

	body := `{"template":"nonexistent"}`
	req := httptest.NewRequest(http.MethodPost, "/api/bundles/clone", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleClone(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestHandleClone_InvalidJSON(t *testing.T) {
	h := NewBundleHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/bundles/clone", bytes.NewBufferString("{bad"))
	w := httptest.NewRecorder()

	h.HandleClone(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleBuild_Success(t *testing.T) {
	h := NewBundleHandler()

	body := `{"name":"Custom","dog_size":"medium","food_type":"dry","extras":["vitamins"],"pack_size":"large"}`
	req := httptest.NewRequest(http.MethodPost, "/api/bundles", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleBuild(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	var b bundle.Bundle
	json.NewDecoder(w.Body).Decode(&b)

	if b.Name != "Custom" {
		t.Errorf("expected name 'Custom', got '%s'", b.Name)
	}
	if b.DogSize != "medium" {
		t.Errorf("expected dog_size 'medium', got '%s'", b.DogSize)
	}
	if b.PackSize != "large" {
		t.Errorf("expected pack_size 'large', got '%s'", b.PackSize)
	}
}

func TestHandleBuild_MissingRequired(t *testing.T) {
	h := NewBundleHandler()

	// Без dog_size — має повернути помилку
	body := `{"name":"Incomplete","food_type":"dry"}`
	req := httptest.NewRequest(http.MethodPost, "/api/bundles", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleBuild(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing dog_size, got %d", w.Code)
	}
}

func TestHandleBuild_DefaultPackSize(t *testing.T) {
	h := NewBundleHandler()

	body := `{"dog_size":"small","food_type":"wet"}`
	req := httptest.NewRequest(http.MethodPost, "/api/bundles", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleBuild(w, req)

	var b bundle.Bundle
	json.NewDecoder(w.Body).Decode(&b)

	if b.PackSize != "standard" {
		t.Errorf("expected default pack_size 'standard', got '%s'", b.PackSize)
	}
}

func TestHandleBuild_MethodNotAllowed(t *testing.T) {
	h := NewBundleHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/bundles", nil)
	w := httptest.NewRecorder()

	h.HandleBuild(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}
