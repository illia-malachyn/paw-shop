// Package handler — HTTP-обробники для нотифікацій та логування.
// Патерни: Facade (notify.NotificationFacade), Proxy (logging.LoggerProxy), Bridge (logging).
package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/illia-malachyn/paw-shop/internal/logging"
	"github.com/illia-malachyn/paw-shop/internal/notify"
)

// NotificationHandler — обробник HTTP-запитів для нотифікацій та логів.
// Використовує NotificationFacade (Facade pattern) та LoggerProxy (Proxy pattern).
type NotificationHandler struct {
	facade *notify.NotificationFacade
	logger *logging.LoggerProxy
}

// NewNotificationHandler — конструктор NotificationHandler.
// Ініціалізує Facade та Proxy з виводом у os.Stdout.
func NewNotificationHandler() *NotificationHandler {
	facade := notify.NewNotificationFacade(os.Stdout, os.Stdout)
	logger := logging.NewLoggerProxy(os.Stdout, "info")
	logger.Log("info", "Notification handler initialized")

	return &NotificationHandler{
		facade: facade,
		logger: logger,
	}
}

// HandleNotify — POST /api/notifications/send
// Відправляє нотифікацію користувачу через Facade.
func (h *NotificationHandler) HandleNotify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID  string `json:"user_id"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.UserID == "" || req.Message == "" {
		http.Error(w, "user_id and message are required", http.StatusBadRequest)
		return
	}

	if err := h.facade.NotifyUser(req.UserID, req.Message); err != nil {
		http.Error(w, "Failed to send notification", http.StatusInternalServerError)
		return
	}

	h.logger.Log("info", fmt.Sprintf("Notification sent to %s", req.UserID))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "sent"})
}

// HandleLogs — GET /api/logs
// Повертає записи логу, відфільтровані за рівнем (query param: level).
func (h *NotificationHandler) HandleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	level := r.URL.Query().Get("level")
	entries := h.logger.GetEntries(level)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

// HandleLogStats — GET /api/logs/stats
// Повертає кількість записів за рівнями та загальну кількість.
func (h *NotificationHandler) HandleLogStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	entries := h.logger.GetEntries("")
	counts := map[string]int{}
	for _, e := range entries {
		counts[e.Level]++
	}
	counts["total"] = h.logger.GetLogCount()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(counts)
}
