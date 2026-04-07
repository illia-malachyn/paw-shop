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
	collection *order.OrderCollection
	validator  order.OrderValidator
}

// NewOrderHandler — створює OrderHandler з тестовими замовленнями та ланцюжком валідаторів.
func NewOrderHandler() *OrderHandler {
	col := order.NewOrderCollection()
	col.Add(order.NewOrder("order-1", []string{"rc-dry-01", "ac-wet-01"}))
	col.Add(order.NewOrder("order-2", []string{"ac-dry-01"}))
	col.Add(order.NewOrder("order-3", []string{"rc-wet-01", "rc-dry-01"}))

	return &OrderHandler{
		collection: col,
		validator:  order.NewValidationChain(),
	}
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
		o, ok := h.collection.GetByID(id)
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

	allOrders := make([]*order.Order, 0, h.collection.Count())
	it := h.collection.CreateIterator()
	for it.HasNext() {
		allOrders = append(allOrders, it.Next())
	}

	reportText := order.GenerateReport(gen, allOrders)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"report_type": reportType,
		"content":     reportText,
	})
}

// HandleStatus — PATCH /api/orders/{id}/status
// Просуває або скасовує замовлення відповідно до State pattern.
func (h *OrderHandler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Strip "/api/orders/" prefix and "/status" suffix to extract order ID
	path := r.URL.Path
	path = strings.TrimPrefix(path, "/api/orders/")
	path = strings.TrimSuffix(path, "/status")
	id := path

	o, ok := h.collection.GetByID(id)
	if !ok {
		http.Error(w, fmt.Sprintf("order not found: %s", id), http.StatusNotFound)
		return
	}

	var req struct {
		Action string `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var err error
	switch req.Action {
	case "next":
		err = o.Next()
	case "cancel":
		err = o.Cancel()
	default:
		http.Error(w, "invalid action: must be \"next\" or \"cancel\"", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":     o.ID,
		"status": o.GetState().Name(),
	})
}

// HandleListOrders — GET /api/orders
// Повертає всі замовлення або фільтровані за статусом (Iterator pattern).
func (h *OrderHandler) HandleListOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := r.URL.Query().Get("status")

	var it order.OrderIterator
	if status != "" {
		it = h.collection.CreateFilteredIterator(status)
	} else {
		it = h.collection.CreateIterator()
	}

	orders := make([]*order.Order, 0)
	for it.HasNext() {
		orders = append(orders, it.Next())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

// HandleCreateOrder — POST /api/orders
// Валідує запит через Chain of Responsibility і створює нове замовлення.
func (h *OrderHandler) HandleCreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Items   []string `json:"items"`
		Address string   `json:"address"`
		Amount  float64  `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	orderReq := order.OrderRequest{
		Items:   req.Items,
		Address: req.Address,
		Amount:  req.Amount,
	}
	if err := h.validator.Validate(&orderReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := fmt.Sprintf("order-%d", h.collection.Count()+1)
	newOrder := order.NewOrder(id, req.Items)
	h.collection.Add(newOrder)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":     newOrder.ID,
		"status": newOrder.GetState().Name(),
		"items":  newOrder.Items,
	})
}

// HandleOrders — маршрутизатор для /api/orders та /api/orders/{id}/status.
func (h *OrderHandler) HandleOrders(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	switch {
	case path == "/api/orders" && r.Method == http.MethodGet:
		h.HandleListOrders(w, r)
	case path == "/api/orders" && r.Method == http.MethodPost:
		h.HandleCreateOrder(w, r)
	case strings.HasSuffix(path, "/status") && r.Method == http.MethodPatch:
		h.HandleStatus(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GetOrders — повертає всі замовлення як map для зворотної сумісності з тестами.
func (h *OrderHandler) GetOrders() map[string]*order.Order {
	result := make(map[string]*order.Order)
	it := h.collection.CreateIterator()
	for it.HasNext() {
		o := it.Next()
		result[o.ID] = o
	}
	return result
}
