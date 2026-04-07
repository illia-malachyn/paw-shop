// Пакет export реалізує патерн Visitor для генерації різних форматів виводу кошика.
// OrderElement — елемент, що приймає відвідувача; OrderVisitor — відвідувач елементів.
// Реалізації: JSONExportVisitor (JSON-формат) та TextReceiptVisitor (текстова квитанція).
package export

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/illia-malachyn/paw-shop/internal/cart"
)

// OrderElement — елемент замовлення, який приймає відвідувача.
type OrderElement interface {
	Accept(visitor OrderVisitor)
}

// OrderVisitor — відвідувач елементів замовлення для генерації різних форматів.
// Distinct method names per element type (Go does not support method overloading).
type OrderVisitor interface {
	VisitItem(item cart.CartItem)
	VisitCart(c cart.Cart)
	Result() string
}

// CartItemElement — обгортка навколо cart.CartItem для участі у патерні Visitor.
type CartItemElement struct {
	Item cart.CartItem
}

// Accept викликає visitor.VisitItem для обробки елемента.
func (e CartItemElement) Accept(visitor OrderVisitor) {
	visitor.VisitItem(e.Item)
}

// CartElement — обгортка навколо cart.Cart для участі у патерні Visitor.
// Accept ітерує всі товари кошика, потім передає сам кошик.
type CartElement struct {
	Cart cart.Cart
}

// Accept ітерує cart.Items, викликаючи VisitItem для кожного, потім викликає VisitCart.
func (e CartElement) Accept(visitor OrderVisitor) {
	for _, item := range e.Cart.Items {
		visitor.VisitItem(item)
	}
	visitor.VisitCart(e.Cart)
}

// jsonItem — внутрішня структура для одного елемента у JSON-виводі.
type jsonItem struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	Subtotal  float64 `json:"subtotal"`
}

// jsonOutput — внутрішня структура для повного JSON-виводу.
type jsonOutput struct {
	Items []jsonItem `json:"items"`
	Total float64    `json:"total"`
}

// JSONExportVisitor — відвідувач, що накопичує елементи та генерує JSON-представлення кошика.
type JSONExportVisitor struct {
	items []jsonItem
	total float64
}

// VisitItem додає елемент до списку та оновлює загальну суму.
func (v *JSONExportVisitor) VisitItem(item cart.CartItem) {
	subtotal := item.Price * float64(item.Quantity)
	v.items = append(v.items, jsonItem{
		ProductID: item.ProductID,
		Name:      item.Name,
		Price:     item.Price,
		Quantity:  item.Quantity,
		Subtotal:  subtotal,
	})
	v.total += subtotal
}

// VisitCart — для JSONExportVisitor загальна сума вже накопичена у VisitItem, тому є no-op.
func (v *JSONExportVisitor) VisitCart(_ cart.Cart) {}

// Result повертає JSON-рядок з усіма елементами та загальною сумою.
func (v *JSONExportVisitor) Result() string {
	items := v.items
	if items == nil {
		items = []jsonItem{}
	}
	out := jsonOutput{Items: items, Total: v.total}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(data)
}

// TextReceiptVisitor — відвідувач, що накопичує рядки та генерує текстову квитанцію.
type TextReceiptVisitor struct {
	lines []string
	total float64
}

// VisitItem додає рядок із деталями товару до квитанції та оновлює загальну суму.
func (v *TextReceiptVisitor) VisitItem(item cart.CartItem) {
	subtotal := item.Price * float64(item.Quantity)
	v.lines = append(v.lines, fmt.Sprintf("%s - %d x %.2f = %.2f", item.Name, item.Quantity, item.Price, subtotal))
	v.total += subtotal
}

// VisitCart додає роздільник і рядок загальної суми до квитанції.
func (v *TextReceiptVisitor) VisitCart(_ cart.Cart) {
	v.lines = append(v.lines, "================")
	v.lines = append(v.lines, fmt.Sprintf("Total: %.2f", v.total))
}

// Result повертає повну текстову квитанцію як один рядок із переносами.
func (v *TextReceiptVisitor) Result() string {
	header := []string{"=== Receipt ==="}
	allLines := append(header, v.lines...)
	return strings.Join(allLines, "\n")
}

// ExportCart — зручна функція для застосування відвідувача до кошика.
// Ітерує cart.Items, викликає VisitItem для кожного, потім VisitCart, повертає Result().
func ExportCart(c cart.Cart, v OrderVisitor) string {
	for _, item := range c.Items {
		v.VisitItem(item)
	}
	v.VisitCart(c)
	return v.Result()
}
