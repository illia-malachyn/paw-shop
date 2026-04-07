package handler

import (
	"encoding/json"
	"net/http"

	"github.com/illia-malachyn/paw-shop/internal/factory"
	"github.com/illia-malachyn/paw-shop/internal/models"
)

// ProductHandler — обробник HTTP-запитів для каталогу товарів.
type ProductHandler struct{}

func NewProductHandler() *ProductHandler {
	return &ProductHandler{}
}

// HandleProducts — GET /api/products
// Повертає каталог товарів, згенерований через абстрактні фабрики брендів.
func (h *ProductHandler) HandleProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	brands := []factory.BrandFactory{
		&factory.RoyalCaninFactory{},
		&factory.AcanaFactory{},
	}

	var catalog []models.ProductResponse

	for _, brand := range brands {
		products := []models.Product{
			brand.CreateDryFood(),
			brand.CreateWetFood(),
			brand.CreateTreat(),
		}

		for _, p := range products {
			catalog = append(catalog, models.ProductResponse{
				ID:       p.GetID(),
				Name:     p.GetName(),
				Price:    p.GetPrice(),
				Category: p.GetCategory(),
				Brand:    brand.BrandName(),
				Details:  p.GetDetails(),
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(catalog)
}
