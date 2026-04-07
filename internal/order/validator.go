package order

import (
	"fmt"
	"strings"
)

// OrderRequest — дані замовлення, що проходять через ланцюжок валідаторів.
type OrderRequest struct {
	Items   []string
	Address string
	Amount  float64
}

// OrderValidator — інтерфейс валідатора замовлення (Chain of Responsibility pattern).
type OrderValidator interface {
	SetNext(validator OrderValidator) OrderValidator
	Validate(req *OrderRequest) error
}

// BaseValidator — базова структура для вбудовування в конкретні валідатори.
// Зберігає посилання на наступний валідатор у ланцюжку.
type BaseValidator struct {
	next OrderValidator
}

// SetNext — встановлює наступний валідатор у ланцюжку, повертає його для fluent chaining.
func (b *BaseValidator) SetNext(v OrderValidator) OrderValidator {
	b.next = v
	return v
}

// passToNext — передає запит наступному валідатору або повертає nil якщо ланцюжок завершено.
func (b *BaseValidator) passToNext(req *OrderRequest) error {
	if b.next != nil {
		return b.next.Validate(req)
	}
	return nil
}

// StockValidator — перевіряє наявність товарів на складі (Chain of Responsibility pattern).
type StockValidator struct {
	BaseValidator
}

// Validate — перевіряє що жоден товар не є "out-of-stock-item".
func (v *StockValidator) Validate(req *OrderRequest) error {
	for _, item := range req.Items {
		if item == "out-of-stock-item" {
			return fmt.Errorf("stock validation failed: item %q is out of stock", item)
		}
	}
	return v.passToNext(req)
}

// AddressValidator — перевіряє наявність адреси доставки (Chain of Responsibility pattern).
type AddressValidator struct {
	BaseValidator
}

// Validate — перевіряє що адреса не є порожньою або пробілами.
func (v *AddressValidator) Validate(req *OrderRequest) error {
	if strings.TrimSpace(req.Address) == "" {
		return fmt.Errorf("address validation failed: address is required")
	}
	return v.passToNext(req)
}

// PaymentValidator — перевіряє що сума оплати більша за нуль (Chain of Responsibility pattern).
type PaymentValidator struct {
	BaseValidator
}

// Validate — перевіряє що сума оплати > 0.
func (v *PaymentValidator) Validate(req *OrderRequest) error {
	if req.Amount <= 0 {
		return fmt.Errorf("payment validation failed: amount must be greater than 0")
	}
	return v.passToNext(req)
}

// NewValidationChain — будує стандартний ланцюжок валідаторів: Stock → Address → Payment.
func NewValidationChain() OrderValidator {
	stock := &StockValidator{}
	address := &AddressValidator{}
	payment := &PaymentValidator{}
	stock.SetNext(address).SetNext(payment)
	return stock
}
