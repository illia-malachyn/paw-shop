package notification

import "fmt"

// --- Observer ---

// PriceObserver — інтерфейс спостерігача за зміною ціни.
type PriceObserver interface {
	OnPriceChanged(productID string, oldPrice, newPrice float64)
}

// PriceSubject — суб'єкт, який зберігає ціни та нотифікує підписників.
type PriceSubject struct {
	prices    map[string]float64
	observers map[string][]PriceObserver // productID -> observers
}

func NewPriceSubject() *PriceSubject {
	return &PriceSubject{
		prices:    make(map[string]float64),
		observers: make(map[string][]PriceObserver),
	}
}

// SetPrice — встановлює ціну і нотифікує спостерігачів якщо вона змінилась.
func (s *PriceSubject) SetPrice(productID string, price float64) {
	oldPrice, exists := s.prices[productID]
	s.prices[productID] = price

	if exists && oldPrice != price {
		for _, obs := range s.observers[productID] {
			obs.OnPriceChanged(productID, oldPrice, price)
		}
	}
}

// GetPrice — повертає поточну ціну товару.
func (s *PriceSubject) GetPrice(productID string) (float64, bool) {
	p, ok := s.prices[productID]
	return p, ok
}

// Subscribe — підписує спостерігача на зміну ціни конкретного товару.
func (s *PriceSubject) Subscribe(productID string, obs PriceObserver) {
	s.observers[productID] = append(s.observers[productID], obs)
}

// --- Конкретні спостерігачі ---

// LogObserver — логує зміну ціни в консоль.
type LogObserver struct {
	UserEmail string
}

func (o *LogObserver) OnPriceChanged(productID string, oldPrice, newPrice float64) {
	fmt.Printf("[NOTIFY] %s: product %s price changed %.2f -> %.2f\n",
		o.UserEmail, productID, oldPrice, newPrice)
}

// NotificationRecord — запис про сповіщення (для тестування та API).
type NotificationRecord struct {
	UserEmail string  `json:"user_email"`
	ProductID string  `json:"product_id"`
	OldPrice  float64 `json:"old_price"`
	NewPrice  float64 `json:"new_price"`
}

// InMemoryObserver — зберігає сповіщення в пам'яті (для API та тестів).
type InMemoryObserver struct {
	UserEmail string
	Records   []NotificationRecord
}

func (o *InMemoryObserver) OnPriceChanged(productID string, oldPrice, newPrice float64) {
	o.Records = append(o.Records, NotificationRecord{
		UserEmail: o.UserEmail,
		ProductID: productID,
		OldPrice:  oldPrice,
		NewPrice:  newPrice,
	})
}
