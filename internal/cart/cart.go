// Пакет cart реалізує патерн Memento для збереження та відновлення стану кошика.
// Originator: Cart, Memento: CartMemento, Caretaker: CartHistory.
package cart

import "fmt"

// CartItem — один елемент кошика: товар із кількістю.
type CartItem struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
}

// Cart — кошик покупок (Originator у патерні Memento).
// Зберігає поточний список товарів та підтримує збереження/відновлення стану.
type Cart struct {
	Items []CartItem
}

// AddItem додає товар до кошика. Якщо товар із таким ProductID вже існує —
// збільшує кількість на item.Quantity.
func (c *Cart) AddItem(item CartItem) {
	for i := range c.Items {
		if c.Items[i].ProductID == item.ProductID {
			c.Items[i].Quantity += item.Quantity
			return
		}
	}
	c.Items = append(c.Items, item)
}

// RemoveItem видаляє товар із кошика за ProductID.
// Повертає помилку, якщо товар не знайдено.
func (c *Cart) RemoveItem(productID string) error {
	for i, item := range c.Items {
		if item.ProductID == productID {
			c.Items = append(c.Items[:i], c.Items[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("товар з product_id %q не знайдено в кошику", productID)
}

// UpdateQuantity змінює кількість товару за ProductID.
// Якщо quantity <= 0, товар видаляється з кошика.
// Повертає помилку, якщо товар не знайдено.
func (c *Cart) UpdateQuantity(productID string, quantity int) error {
	for i, item := range c.Items {
		if item.ProductID == productID {
			if quantity <= 0 {
				c.Items = append(c.Items[:i], c.Items[i+1:]...)
				return nil
			}
			c.Items[i].Quantity = quantity
			return nil
		}
	}
	return fmt.Errorf("товар з product_id %q не знайдено в кошику", productID)
}

// Save створює знімок поточного стану кошика (CartMemento).
// Повертає значення (не вказівник) — незмінний знімок.
func (c *Cart) Save() CartMemento {
	// deep copy — not assignment
	snapshot := make([]CartItem, len(c.Items))
	copy(snapshot, c.Items)
	return CartMemento{items: snapshot}
}

// Restore відновлює стан кошика з CartMemento.
func (c *Cart) Restore(m CartMemento) {
	// deep copy — not assignment
	c.Items = make([]CartItem, len(m.items))
	copy(c.Items, m.items)
}

// CartMemento — непрозорий знімок стану кошика (Memento у патерні Memento).
// Поле items є невідкритим: доглядач не може змінити внутрішній стан.
type CartMemento struct {
	items []CartItem
}

// CartHistory — стек мементо для підтримки undo (Caretaker у патерні Memento).
type CartHistory struct {
	stack []CartMemento
}

// Push додає мементо до стека.
func (h *CartHistory) Push(m CartMemento) {
	h.stack = append(h.stack, m)
}

// Pop повертає останній мементо та видаляє його зі стека (LIFO).
// Повертає false, якщо стек порожній.
func (h *CartHistory) Pop() (CartMemento, bool) {
	if len(h.stack) == 0 {
		return CartMemento{}, false
	}
	last := h.stack[len(h.stack)-1]
	h.stack = h.stack[:len(h.stack)-1]
	return last, true
}
