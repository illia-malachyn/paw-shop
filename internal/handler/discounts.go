package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/illia-malachyn/paw-shop/internal/discount"
	"github.com/illia-malachyn/paw-shop/internal/notification"
)

// DiscountHandler — обробник HTTP-запитів для знижок та підписок.
type DiscountHandler struct {
	subject *notification.PriceSubject
	history *discount.CommandHistory
	// in-memory observers для API відповідей
	observers map[string]*notification.InMemoryObserver
}

func NewDiscountHandler() *DiscountHandler {
	subject := notification.NewPriceSubject()

	// Ініціалізуємо ціни товарів
	subject.SetPrice("rc-dry-01", 1450)
	subject.SetPrice("rc-wet-01", 520)
	subject.SetPrice("ac-dry-01", 1850)
	subject.SetPrice("ac-wet-01", 680)

	return &DiscountHandler{
		subject:   subject,
		history:   discount.NewCommandHistory(),
		observers: make(map[string]*notification.InMemoryObserver),
	}
}

// HandleApply — POST /api/discounts/apply
func (h *DiscountHandler) HandleApply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ProductID    string  `json:"product_id"`
		DiscountType string  `json:"discount_type"` // percent, fixed, buy_n_get_one
		Value        float64 `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	strategy := resolveStrategy(req.DiscountType, req.Value)
	if strategy == nil {
		http.Error(w, "Unknown discount type", http.StatusBadRequest)
		return
	}

	cmd := &discount.ApplyDiscountCommand{
		ProductID: req.ProductID,
		Strategy:  strategy,
		Subject:   h.subject,
	}

	newPrice := h.history.Execute(cmd)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"product_id": req.ProductID,
		"new_price":  newPrice,
		"discount":   strategy.Name(),
	})
}

// HandleUndo — POST /api/discounts/undo
func (h *DiscountHandler) HandleUndo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !h.history.HasHistory() {
		http.Error(w, "Nothing to undo", http.StatusBadRequest)
		return
	}

	restoredPrice := h.history.Undo()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"restored_price": restoredPrice,
	})
}

// HandleSubscribe — POST /api/products/{id}/subscribe
func (h *DiscountHandler) HandleSubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Парсимо product ID з URL: /api/products/rc-dry-01/subscribe
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	productID := parts[3]

	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	obs := &notification.InMemoryObserver{UserEmail: req.Email}
	h.subject.Subscribe(productID, obs)
	h.subject.Subscribe(productID, &notification.LogObserver{UserEmail: req.Email})
	h.observers[req.Email] = obs

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"subscribed": true,
		"product_id": productID,
		"email":      req.Email,
	})
}

func resolveStrategy(discountType string, value float64) discount.Strategy {
	switch discountType {
	case "percent":
		return &discount.PercentStrategy{Percent: value}
	case "fixed":
		return &discount.FixedStrategy{Amount: value}
	case "buy_n_get_one":
		return &discount.BuyNGetOneStrategy{N: int(value)}
	default:
		return nil
	}
}
