package handler

import (
	"encoding/json"
	"net/http"

	"github.com/illia-malachyn/paw-shop/internal/bundle"
)

// BundleHandler — обробник HTTP-запитів для наборів корму.
type BundleHandler struct {
	registry *bundle.BundleRegistry
}

func NewBundleHandler() *BundleHandler {
	return &BundleHandler{
		registry: bundle.NewBundleRegistry(),
	}
}

// HandleTemplates — GET /api/bundles/templates
// Повертає список доступних шаблонів.
func (h *BundleHandler) HandleTemplates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.registry.List())
}

// HandleClone — POST /api/bundles/clone
// Клонує шаблон і застосовує зміни від користувача.
func (h *BundleHandler) HandleClone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Template string   `json:"template"` // ключ шаблону: "puppy", "large_breed", "senior"
		Name     string   `json:"name"`     // нова назва (опціонально)
		Extras   []string `json:"extras"`   // замінити extras (опціонально)
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	b := h.registry.Get(req.Template)
	if b == nil {
		http.Error(w, "Template not found", http.StatusNotFound)
		return
	}

	if req.Name != "" {
		b.Name = req.Name
	}
	if req.Extras != nil {
		b.Extras = req.Extras
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(b)
}

// HandleBuild — POST /api/bundles
// Створює набір з нуля через Builder.
func (h *BundleHandler) HandleBuild(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name     string   `json:"name"`
		DogSize  string   `json:"dog_size"`
		FoodType string   `json:"food_type"`
		Extras   []string `json:"extras"`
		PackSize string   `json:"pack_size"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	builder := bundle.NewBundleBuilder().
		SetName(req.Name).
		SetDogSize(req.DogSize).
		SetFoodType(req.FoodType).
		SetPackSize(req.PackSize)

	for _, extra := range req.Extras {
		builder.AddExtra(extra)
	}

	b, err := builder.Build()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(b)
}
