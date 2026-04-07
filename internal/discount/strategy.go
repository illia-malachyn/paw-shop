package discount

// --- Strategy ---

// Strategy — інтерфейс стратегії знижки.
// Кожна реалізація обчислює фінальну ціну по-своєму.
type Strategy interface {
	Name() string
	Apply(price float64) float64
}

// PercentStrategy — відсоткова знижка (наприклад, 10%).
type PercentStrategy struct {
	Percent float64
}

func (s *PercentStrategy) Name() string { return "percent" }
func (s *PercentStrategy) Apply(price float64) float64 {
	return price * (1 - s.Percent/100)
}

// FixedStrategy — фіксована знижка в грн (наприклад, -50 грн).
type FixedStrategy struct {
	Amount float64
}

func (s *FixedStrategy) Name() string { return "fixed" }
func (s *FixedStrategy) Apply(price float64) float64 {
	result := price - s.Amount
	if result < 0 {
		return 0
	}
	return result
}

// BuyNGetOneStrategy — акція "купи N, отримай 1 безкоштовно".
// Розраховує ціну за (quantity) одиниць товару.
type BuyNGetOneStrategy struct {
	N int // купи N штук
}

func (s *BuyNGetOneStrategy) Name() string { return "buy_n_get_one" }
func (s *BuyNGetOneStrategy) Apply(price float64) float64 {
	// Ціна за одиницю зі знижкою: платиш за N із N+1
	paid := float64(s.N)
	total := float64(s.N + 1)
	return price * (paid / total)
}
