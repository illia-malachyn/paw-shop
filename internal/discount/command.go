package discount

import "github.com/illia-malachyn/paw-shop/internal/notification"

// --- Command ---

// Command — інтерфейс команди із підтримкою скасування.
type Command interface {
	Execute() float64
	Undo() float64
}

// ApplyDiscountCommand — команда застосування знижки до товару.
// Зберігає стару ціну для можливості скасування.
type ApplyDiscountCommand struct {
	ProductID string
	Strategy  Strategy
	Subject   *notification.PriceSubject
	oldPrice  float64
	newPrice  float64
}

func (c *ApplyDiscountCommand) Execute() float64 {
	price, ok := c.Subject.GetPrice(c.ProductID)
	if !ok {
		return 0
	}
	c.oldPrice = price
	c.newPrice = c.Strategy.Apply(price)
	c.Subject.SetPrice(c.ProductID, c.newPrice)
	return c.newPrice
}

func (c *ApplyDiscountCommand) Undo() float64 {
	c.Subject.SetPrice(c.ProductID, c.oldPrice)
	return c.oldPrice
}

// CommandHistory — зберігає історію виконаних команд для undo.
type CommandHistory struct {
	history []Command
}

func NewCommandHistory() *CommandHistory {
	return &CommandHistory{}
}

func (h *CommandHistory) Execute(cmd Command) float64 {
	result := cmd.Execute()
	h.history = append(h.history, cmd)
	return result
}

func (h *CommandHistory) Undo() float64 {
	if len(h.history) == 0 {
		return 0
	}
	last := h.history[len(h.history)-1]
	h.history = h.history[:len(h.history)-1]
	return last.Undo()
}

func (h *CommandHistory) HasHistory() bool {
	return len(h.history) > 0
}
