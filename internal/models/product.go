package models

// Product — базовий інтерфейс для всіх товарів магазину.
type Product interface {
	GetID() string
	GetName() string
	GetPrice() float64
	GetCategory() string
	GetDetails() map[string]interface{}
}

// DryFood — сухий корм з вагою упаковки.
type DryFood struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Category string  `json:"category"`
	WeightKg float64 `json:"weight_kg"`
	Flavor   string  `json:"flavor"`
}

func (d *DryFood) GetID() string       { return d.ID }
func (d *DryFood) GetName() string     { return d.Name }
func (d *DryFood) GetPrice() float64   { return d.Price }
func (d *DryFood) GetCategory() string { return d.Category }
func (d *DryFood) GetDetails() map[string]interface{} {
	return map[string]interface{}{
		"weight_kg": d.WeightKg,
		"flavor":    d.Flavor,
	}
}

// WetFood — вологий корм з кількістю банок.
type WetFood struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Category string  `json:"category"`
	CanCount int     `json:"can_count"`
	Flavor   string  `json:"flavor"`
}

func (w *WetFood) GetID() string       { return w.ID }
func (w *WetFood) GetName() string     { return w.Name }
func (w *WetFood) GetPrice() float64   { return w.Price }
func (w *WetFood) GetCategory() string { return w.Category }
func (w *WetFood) GetDetails() map[string]interface{} {
	return map[string]interface{}{
		"can_count": w.CanCount,
		"flavor":    w.Flavor,
	}
}

// Treat — ласощі з кількістю штук в упаковці.
type Treat struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Category string  `json:"category"`
	PieceCount int   `json:"piece_count"`
	Type     string  `json:"type"` // dental, training, chew
}

func (t *Treat) GetID() string       { return t.ID }
func (t *Treat) GetName() string     { return t.Name }
func (t *Treat) GetPrice() float64   { return t.Price }
func (t *Treat) GetCategory() string { return t.Category }
func (t *Treat) GetDetails() map[string]interface{} {
	return map[string]interface{}{
		"piece_count": t.PieceCount,
		"type":        t.Type,
	}
}

// ProductResponse — структура для JSON-відповіді API.
type ProductResponse struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Price    float64                `json:"price"`
	Category string                 `json:"category"`
	Brand    string                 `json:"brand"`
	Details  map[string]interface{} `json:"details"`
}
