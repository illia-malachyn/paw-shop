package factory

import "github.com/illia-malachyn/paw-shop/internal/models"

// --- Abstract Factory ---

// BrandFactory — абстрактна фабрика, що створює сімейство продуктів одного бренду.
// Кожен бренд має свій сухий корм, вологий корм та ласощі.
type BrandFactory interface {
	CreateDryFood() models.Product
	CreateWetFood() models.Product
	CreateTreat() models.Product
	BrandName() string
}

// RoyalCaninFactory — фабрика продуктів бренду Royal Canin.
type RoyalCaninFactory struct{}

func (f *RoyalCaninFactory) BrandName() string { return "Royal Canin" }

func (f *RoyalCaninFactory) CreateDryFood() models.Product {
	return &models.DryFood{
		ID:       "rc-dry-01",
		Name:     "Royal Canin Maxi Adult",
		Price:    1450,
		Category: "dry",
		WeightKg: 15,
		Flavor:   "chicken",
	}
}

func (f *RoyalCaninFactory) CreateWetFood() models.Product {
	return &models.WetFood{
		ID:       "rc-wet-01",
		Name:     "Royal Canin Chunks in Gravy",
		Price:    520,
		Category: "wet",
		CanCount: 6,
		Flavor:   "beef",
	}
}

func (f *RoyalCaninFactory) CreateTreat() models.Product {
	return &models.Treat{
		ID:         "rc-treat-01",
		Name:       "Royal Canin Dental Sticks",
		Price:      280,
		Category:   "treat",
		PieceCount: 7,
		Type:       "dental",
	}
}

// AcanaFactory — фабрика продуктів бренду Acana.
type AcanaFactory struct{}

func (f *AcanaFactory) BrandName() string { return "Acana" }

func (f *AcanaFactory) CreateDryFood() models.Product {
	return &models.DryFood{
		ID:       "ac-dry-01",
		Name:     "Acana Wild Prairie",
		Price:    1850,
		Category: "dry",
		WeightKg: 11.4,
		Flavor:   "poultry",
	}
}

func (f *AcanaFactory) CreateWetFood() models.Product {
	return &models.WetFood{
		ID:       "ac-wet-01",
		Name:     "Acana Premium Pate",
		Price:    680,
		Category: "wet",
		CanCount: 12,
		Flavor:   "lamb",
	}
}

func (f *AcanaFactory) CreateTreat() models.Product {
	return &models.Treat{
		ID:         "ac-treat-01",
		Name:       "Acana Crunchy Biscuits",
		Price:      350,
		Category:   "treat",
		PieceCount: 20,
		Type:       "training",
	}
}

// GetBrandFactory — повертає фабрику бренду за назвою.
func GetBrandFactory(brand string) BrandFactory {
	switch brand {
	case "royal_canin":
		return &RoyalCaninFactory{}
	case "acana":
		return &AcanaFactory{}
	default:
		return nil
	}
}
