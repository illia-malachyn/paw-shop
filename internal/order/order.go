package order

import "fmt"

// Order — замовлення в магазині.
type Order struct {
	ID     string
	Status string // "new", "confirmed", "rejected"
	Items  []string
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
