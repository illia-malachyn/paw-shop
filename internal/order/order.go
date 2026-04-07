package order

import "fmt"

// Order — замовлення в магазині.
type Order struct {
	ID     string   `json:"id"`
	Status string   `json:"status"` // синхронізований з state.Name()
	Items  []string `json:"items"`
	state  OrderState
}

// NewOrder — створює нове замовлення у початковому стані "new".
func NewOrder(id string, items []string) *Order {
	return &Order{
		ID:     id,
		Items:  items,
		state:  &NewState{},
		Status: "new",
	}
}

// Next — переводить замовлення до наступного стану (State pattern).
func (o *Order) Next() error {
	return o.state.Next(o)
}

// Cancel — скасовує замовлення, якщо це допускається поточним станом.
func (o *Order) Cancel() error {
	return o.state.Cancel(o)
}

// GetState — повертає поточний стан замовлення.
func (o *Order) GetState() OrderState {
	return o.state
}

// OrderCommand — інтерфейс команди для замовлення.
type OrderCommand interface {
	Execute() error
}

// ConfirmOrderCommand — команда підтвердження замовлення (Command pattern).
type ConfirmOrderCommand struct {
	Order *Order
}

func (c *ConfirmOrderCommand) Execute() error {
	if c.Order.Status != "new" {
		return fmt.Errorf("cannot confirm order %s: status is %q, expected \"new\"", c.Order.ID, c.Order.Status)
	}
	c.Order.Status = "confirmed"
	return nil
}

// RejectOrderCommand — команда відхилення замовлення (Command pattern).
type RejectOrderCommand struct {
	Order *Order
}

func (c *RejectOrderCommand) Execute() error {
	if c.Order.Status != "new" {
		return fmt.Errorf("cannot reject order %s: status is %q, expected \"new\"", c.Order.ID, c.Order.Status)
	}
	c.Order.Status = "rejected"
	return nil
}

// MacroCommand — команда, що об'єднує кілька команд і виконує їх послідовно (MacroCommand pattern).
type MacroCommand struct {
	commands []OrderCommand
}

// NewMacroCommand — створює MacroCommand з переданих команд.
func NewMacroCommand(commands []OrderCommand) *MacroCommand {
	return &MacroCommand{commands: commands}
}

// Execute — виконує всі команди послідовно. Повертає помилку і зупиняється при першій невдачі.
func (m *MacroCommand) Execute() error {
	for _, cmd := range m.commands {
		if err := cmd.Execute(); err != nil {
			return err
		}
	}
	return nil
}
