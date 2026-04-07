package factory

import "github.com/illia-malachyn/paw-shop/internal/models"

// --- Factory Method ---

// ProductFactory — інтерфейс фабричного методу.
// Кожна конкретна фабрика створює продукт своєї категорії.
type ProductFactory interface {
	Create(id, name string, price float64, opts map[string]interface{}) models.Product
}

// DryFoodFactory — фабрика для створення сухого корму.
type DryFoodFactory struct{}

func (f *DryFoodFactory) Create(id, name string, price float64, opts map[string]interface{}) models.Product {
	weightKg := 1.0
	if v, ok := opts["weight_kg"].(float64); ok {
		weightKg = v
	}
	flavor := "chicken"
	if v, ok := opts["flavor"].(string); ok {
		flavor = v
	}
	return &models.DryFood{
		ID:       id,
		Name:     name,
		Price:    price,
		Category: "dry",
		WeightKg: weightKg,
		Flavor:   flavor,
	}
}

// WetFoodFactory — фабрика для створення вологого корму.
type WetFoodFactory struct{}

func (f *WetFoodFactory) Create(id, name string, price float64, opts map[string]interface{}) models.Product {
	canCount := 6
	if v, ok := opts["can_count"].(int); ok {
		canCount = v
	}
	flavor := "beef"
	if v, ok := opts["flavor"].(string); ok {
		flavor = v
	}
	return &models.WetFood{
		ID:       id,
		Name:     name,
		Price:    price,
		Category: "wet",
		CanCount: canCount,
		Flavor:   flavor,
	}
}

// TreatFactory — фабрика для створення ласощів.
type TreatFactory struct{}

func (f *TreatFactory) Create(id, name string, price float64, opts map[string]interface{}) models.Product {
	pieceCount := 10
	if v, ok := opts["piece_count"].(int); ok {
		pieceCount = v
	}
	treatType := "chew"
	if v, ok := opts["type"].(string); ok {
		treatType = v
	}
	return &models.Treat{
		ID:         id,
		Name:       name,
		Price:      price,
		Category:   "treat",
		PieceCount: pieceCount,
		Type:       treatType,
	}
}

// GetFactory — повертає потрібну фабрику за назвою категорії.
func GetFactory(category string) ProductFactory {
	switch category {
	case "dry":
		return &DryFoodFactory{}
	case "wet":
		return &WetFoodFactory{}
	case "treat":
		return &TreatFactory{}
	default:
		return nil
	}
}
