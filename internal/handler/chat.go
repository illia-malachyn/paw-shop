package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/illia-malachyn/paw-shop/internal/chat"
)

// ChatHandler — обробник HTTP-запитів для чату підтримки (Mediator pattern).
type ChatHandler struct {
	mediator *chat.SupportChatMediator
}

// NewChatHandler — конструктор ChatHandler з попередньо зареєстрованими учасниками.
func NewChatHandler() *ChatHandler {
	mediator := chat.NewSupportChatMediator()

	customer := chat.NewCustomer("customer1", mediator)
	manager := chat.NewManager("manager1", mediator)

	mediator.AddParticipant(customer)
	mediator.AddParticipant(manager)

	return &ChatHandler{mediator: mediator}
}

// HandleSend — POST /api/chat/send
// Надсилає повідомлення від одного учасника до іншого через медіатора.
func (h *ChatHandler) HandleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		From    string `json:"from"`
		To      string `json:"to"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.From == "" || req.To == "" || req.Message == "" {
		http.Error(w, "missing required fields: from, to, message", http.StatusBadRequest)
		return
	}

	h.mediator.SendMessage(req.From, req.To, req.Message)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "sent",
		"from":   req.From,
		"to":     req.To,
	})
}

// HandleHistory — GET /api/chat/history?participant=<name>
// Повертає історію повідомлень для вказаного учасника.
func (h *ChatHandler) HandleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	participant := r.URL.Query().Get("participant")
	if participant == "" {
		http.Error(w, "missing query parameter: participant", http.StatusBadRequest)
		return
	}

	history := h.mediator.GetHistory(participant)
	if history == nil {
		history = []chat.Message{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

// HandleChat — маршрутизатор для /api/chat/send та /api/chat/history.
func (h *ChatHandler) HandleChat(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	switch {
	case strings.HasSuffix(path, "/send") && r.Method == http.MethodPost:
		h.HandleSend(w, r)
	case strings.HasSuffix(path, "/history") && r.Method == http.MethodGet:
		h.HandleHistory(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
