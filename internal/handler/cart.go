// Пакет handler — обробники HTTP-запитів для кошика покупок.
// CartHandler реалізує патерни Memento (збереження/відновлення стану) та Visitor (експорт кошика).
package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/illia-malachyn/paw-shop/internal/cart"
	"github.com/illia-malachyn/paw-shop/internal/export"
)

// CartHandler — обробник HTTP-запитів для кошика покупок.
// Використовує Memento для undo та Visitor для експорту.
type CartHandler struct {
	cart    *cart.Cart
	history *cart.CartHistory
}

// NewCartHandler створює CartHandler з порожнім кошиком та порожньою історією.
func NewCartHandler() *CartHandler {
	return &CartHandler{
		cart:    &cart.Cart{},
		history: &cart.CartHistory{},
	}
}

// HandleCart — маршрутизатор для /api/cart та /api/cart/*.
// GET  /api/cart              — поточний вміст кошика
// POST /api/cart/add          — додати товар
// POST /api/cart/remove       — видалити товар
// POST /api/cart/undo         — скасувати останню зміну (Memento)
// GET  /api/cart/export       — експорт кошика (Visitor)
func (h *CartHandler) HandleCart(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	switch {
	case path == "/api/cart" && r.Method == http.MethodGet:
		h.handleGet(w, r)
	case strings.HasSuffix(path, "/api/cart/add") && r.Method == http.MethodPost:
		h.handleAdd(w, r)
	case path == "/api/cart/add":
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	case strings.HasSuffix(path, "/api/cart/remove") && r.Method == http.MethodPost:
		h.handleRemove(w, r)
	case path == "/api/cart/remove":
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	case strings.HasSuffix(path, "/api/cart/undo") && r.Method == http.MethodPost:
		h.handleUndo(w, r)
	case path == "/api/cart/undo":
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	case strings.HasSuffix(path, "/api/cart/export") && r.Method == http.MethodGet:
		h.handleExport(w, r)
	case path == "/api/cart/export":
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAdd — POST /api/cart/add
// Декодує CartItem, валідує поля, зберігає стан у Memento та додає товар до кошика.
func (h *CartHandler) handleAdd(w http.ResponseWriter, r *http.Request) {
	var item cart.CartItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if item.ProductID == "" || item.Name == "" || item.Price <= 0 || item.Quantity <= 0 {
		http.Error(w, "product_id and name must be non-empty, price and quantity must be > 0", http.StatusBadRequest)
		return
	}

	// Зберігаємо стан перед мутацією (Memento)
	h.history.Push(h.cart.Save())
	h.cart.AddItem(item)

	items := h.cart.Items
	if items == nil {
		items = []cart.CartItem{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "added",
		"cart":   items,
	})
}

// handleRemove — POST /api/cart/remove
// Декодує product_id, зберігає стан у Memento та видаляє товар із кошика.
func (h *CartHandler) handleRemove(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProductID string `json:"product_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Зберігаємо стан перед мутацією (Memento)
	h.history.Push(h.cart.Save())

	if err := h.cart.RemoveItem(req.ProductID); err != nil {
		// Відновлюємо стан — мутація не відбулась
		if m, ok := h.history.Pop(); ok {
			h.cart.Restore(m)
		}
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	items := h.cart.Items
	if items == nil {
		items = []cart.CartItem{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "removed",
		"cart":   items,
	})
}

// handleUndo — POST /api/cart/undo
// Відновлює попередній стан кошика з CartHistory (Memento).
func (h *CartHandler) handleUndo(w http.ResponseWriter, r *http.Request) {
	m, ok := h.history.Pop()
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "nothing to undo",
		})
		return
	}

	h.cart.Restore(m)

	items := h.cart.Items
	if items == nil {
		items = []cart.CartItem{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "undone",
		"cart":   items,
	})
}

// handleGet — GET /api/cart
// Повертає поточний вміст кошика як JSON-масив.
func (h *CartHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	items := h.cart.Items
	if items == nil {
		items = []cart.CartItem{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// handleExport — GET /api/cart/export?format=json|text
// Застосовує відповідний Visitor для генерації експорту кошика.
func (h *CartHandler) handleExport(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")

	switch format {
	case "json":
		v := &export.JSONExportVisitor{}
		result := export.ExportCart(*h.cart, v)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(result))
	case "text":
		v := &export.TextReceiptVisitor{}
		result := export.ExportCart(*h.cart, v)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(result))
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "format must be json or text",
		})
	}
}
