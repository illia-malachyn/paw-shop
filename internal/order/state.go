package order

import "fmt"

// OrderState — інтерфейс стану замовлення (State pattern).
// Кожна реалізація визначає допустимі переходи з поточного стану.
type OrderState interface {
	Name() string
	Next(order *Order) error
	Cancel(order *Order) error
}

// NewState — початковий стан замовлення.
type NewState struct{}

// Name — повертає назву стану.
func (s *NewState) Name() string { return "new" }

// Next — переводить замовлення у стан "confirmed".
func (s *NewState) Next(o *Order) error {
	o.state = &ConfirmedState{}
	o.Status = "confirmed"
	return nil
}

// Cancel — скасовує замовлення, переводячи у стан "cancelled".
func (s *NewState) Cancel(o *Order) error {
	o.state = &CancelledState{}
	o.Status = "cancelled"
	return nil
}

// ConfirmedState — стан підтвердженого замовлення.
type ConfirmedState struct{}

// Name — повертає назву стану.
func (s *ConfirmedState) Name() string { return "confirmed" }

// Next — переводить замовлення у стан "shipped".
func (s *ConfirmedState) Next(o *Order) error {
	o.state = &ShippedState{}
	o.Status = "shipped"
	return nil
}

// Cancel — скасовує підтверджене замовлення.
func (s *ConfirmedState) Cancel(o *Order) error {
	o.state = &CancelledState{}
	o.Status = "cancelled"
	return nil
}

// ShippedState — стан відправленого замовлення.
type ShippedState struct{}

// Name — повертає назву стану.
func (s *ShippedState) Name() string { return "shipped" }

// Next — переводить замовлення у стан "delivered".
func (s *ShippedState) Next(o *Order) error {
	o.state = &DeliveredState{}
	o.Status = "delivered"
	return nil
}

// Cancel — повертає помилку: відправлене замовлення не можна скасувати.
func (s *ShippedState) Cancel(o *Order) error {
	return fmt.Errorf("cannot cancel order: already shipped")
}

// DeliveredState — стан доставленого замовлення.
type DeliveredState struct{}

// Name — повертає назву стану.
func (s *DeliveredState) Name() string { return "delivered" }

// Next — повертає помилку: доставлене замовлення не має наступного стану.
func (s *DeliveredState) Next(o *Order) error {
	return fmt.Errorf("order already delivered")
}

// Cancel — повертає помилку: доставлене замовлення не можна скасувати.
func (s *DeliveredState) Cancel(o *Order) error {
	return fmt.Errorf("cannot cancel order: already delivered")
}

// CancelledState — стан скасованого замовлення.
type CancelledState struct{}

// Name — повертає назву стану.
func (s *CancelledState) Name() string { return "cancelled" }

// Next — повертає помилку: скасоване замовлення не може бути просунуто далі.
func (s *CancelledState) Next(o *Order) error {
	return fmt.Errorf("order is cancelled")
}

// Cancel — повертає помилку: замовлення вже скасовано.
func (s *CancelledState) Cancel(o *Order) error {
	return fmt.Errorf("order is already cancelled")
}
