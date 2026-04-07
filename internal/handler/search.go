package handler

import (
	"encoding/json"
	"net/http"

	"github.com/illia-malachyn/paw-shop/internal/factory"
	"github.com/illia-malachyn/paw-shop/internal/models"
	"github.com/illia-malachyn/paw-shop/internal/search"
)

// SearchHandler — обробник HTTP-запитів для пошуку товарів за запитом (Interpreter pattern).
type SearchHandler struct{}

// NewSearchHandler — конструктор SearchHandler.
func NewSearchHandler() *SearchHandler {
	return &SearchHandler{}
}

// HandleSearch — GET /api/products/search?q=<query>
// Розбирає пошуковий запит через Interpreter і повертає відфільтровані товари.
func (h *SearchHandler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, "missing query parameter: q", http.StatusBadRequest)
		return
	}

	expr, err := search.Parse(q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	brands := []factory.BrandFactory{
		&factory.RoyalCaninFactory{},
		&factory.AcanaFactory{},
	}

	var matches []models.ProductResponse

	for _, brand := range brands {
		products := []models.Product{
			brand.CreateDryFood(),
			brand.CreateWetFood(),
			brand.CreateTreat(),
		}

		for _, p := range products {
			pd := search.ProductData{
				ID:       p.GetID(),
				Name:     p.GetName(),
				Brand:    brand.BrandName(),
				Category: p.GetCategory(),
				Price:    p.GetPrice(),
			}

			if expr.Interpret(pd) {
				matches = append(matches, models.ProductResponse{
					ID:       p.GetID(),
					Name:     p.GetName(),
					Price:    p.GetPrice(),
					Category: p.GetCategory(),
					Brand:    brand.BrandName(),
					Details:  p.GetDetails(),
				})
			}
		}
	}

	if matches == nil {
		matches = []models.ProductResponse{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matches)
}
