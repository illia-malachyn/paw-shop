package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/illia-malachyn/paw-shop/internal/order"
)

// OrderHandler — обробник HTTP-запитів для замовлень та звітів.
type OrderHandler struct {
	orders map[string]*order.Order
}

// NewOrderHandler — створює OrderHandler з тестовими замовленнями.
func NewOrderHandler() *OrderHandler {
	orders := map[string]*order.Order{
		"order-1": {ID: "order-1", Status: "new", Items: []string{"rc-dry-01", "ac-wet-01"}},
		"order-2": {ID: "order-2", Status: "new", Items: []string{"ac-dry-01"}},
		"order-3": {ID: "order-3", Status: "new", Items: []string{"rc-wet-01", "rc-dry-01"}},
	}
	return &OrderHandler{orders: orders}
}

// HandleBatch — POST /api/orders/batch
// Приймає order_ids та action ("confirm"|"reject"), застосовує MacroCommand.
func (h *OrderHandler) HandleBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		OrderIDs []string `json:"order_ids"`
		Action   string   `json:"action"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Action != "confirm" && req.Action != "reject" {
		http.Error(w, "Invalid action: must be \"confirm\" or \"reject\"", http.StatusBadRequest)
		return
	}

	commands := make([]order.OrderCommand, 0, len(req.OrderIDs))
	for _, id := range req.OrderIDs {
		o, ok := h.orders[id]
		if !ok {
			http.Error(w, fmt.Sprintf("order not found: %s", id), http.StatusBadRequest)
			return
		}
		switch req.Action {
		case "confirm":
			commands = append(commands, &order.ConfirmOrderCommand{Order: o})
		case "reject":
			commands = append(commands, &order.RejectOrderCommand{Order: o})
		}
	}

	macro := order.NewMacroCommand(commands)
	if err := macro.Execute(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"processed": len(req.OrderIDs),
		"action":    req.Action,
	})
}

// HandleReport — GET /api/reports/{type}
// type = "daily" або "weekly". Генерує звіт за допомогою Template Method.
func (h *OrderHandler) HandleReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	reportType := strings.TrimPrefix(r.URL.Path, "/api/reports/")

	var gen order.ReportGenerator
	switch reportType {
	case "daily":
		gen = &order.DailyReportGenerator{}
	case "weekly":
		gen = &order.WeeklyReportGenerator{}
	default:
		http.Error(w, "unknown report type", http.StatusBadRequest)
		return
	}

	allOrders := make([]*order.Order, 0, len(h.orders))
	for _, o := range h.orders {
		allOrders = append(allOrders, o)
	}

	reportText := order.GenerateReport(gen, allOrders)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"report_type": reportType,
		"content":     reportText,
	})
}

// GetOrders — повертає всі замовлення (для тестів та звітів).
func (h *OrderHandler) GetOrders() map[string]*order.Order {
	return h.orders
}
