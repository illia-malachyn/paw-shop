// Package chat реалізує патерн Mediator для чату підтримки.
// Customer та Manager спілкуються виключно через SupportChatMediator,
// не маючи прямих посилань один на одного.
package chat

// Message — повідомлення між учасниками чату.
type Message struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Content string `json:"content"`
}

// ChatMediator — інтерфейс медіатора, що координує обмін повідомленнями між учасниками.
type ChatMediator interface {
	SendMessage(from, to, message string)
	AddParticipant(participant ChatParticipant)
}

// ChatParticipant — інтерфейс учасника чату.
type ChatParticipant interface {
	GetName() string
	Receive(from, message string)
}

// SupportChatMediator — медіатор патерну Mediator для чату підтримки.
// Координує повідомлення між Customer та Manager, зберігає історію.
type SupportChatMediator struct {
	participants map[string]ChatParticipant
	history      []Message
}

// NewSupportChatMediator — конструктор SupportChatMediator.
func NewSupportChatMediator() *SupportChatMediator {
	return &SupportChatMediator{
		participants: make(map[string]ChatParticipant),
		history:      []Message{},
	}
}

// AddParticipant — реєструє учасника в медіаторі за іменем.
func (m *SupportChatMediator) AddParticipant(p ChatParticipant) {
	m.participants[p.GetName()] = p
}

// SendMessage — надсилає повідомлення від одного учасника до іншого через медіатора.
// Зберігає повідомлення в історії. Якщо отримувач не знайдений — тихо ігнорує.
func (m *SupportChatMediator) SendMessage(from, to, message string) {
	m.history = append(m.history, Message{From: from, To: to, Content: message})

	recipient, ok := m.participants[to]
	if !ok {
		return
	}
	recipient.Receive(from, message)
}

// GetHistory — повертає всі повідомлення, де учасник є відправником або отримувачем.
func (m *SupportChatMediator) GetHistory(participant string) []Message {
	var result []Message
	for _, msg := range m.history {
		if msg.From == participant || msg.To == participant {
			result = append(result, msg)
		}
	}
	return result
}

// --- Конкретні учасники ---

// Customer — клієнт, учасник чату підтримки.
// Зберігає посилання тільки на медіатора, не на інших учасників.
type Customer struct {
	name     string
	mediator ChatMediator
	received []Message
}

// NewCustomer — конструктор Customer.
func NewCustomer(name string, mediator ChatMediator) *Customer {
	return &Customer{
		name:     name,
		mediator: mediator,
		received: []Message{},
	}
}

// GetName — повертає ім'я клієнта.
func (c *Customer) GetName() string {
	return c.name
}

// Receive — отримує повідомлення від іншого учасника через медіатора.
func (c *Customer) Receive(from, message string) {
	c.received = append(c.received, Message{From: from, To: c.name, Content: message})
}

// GetReceived — повертає всі отримані повідомлення (для тестів та хендлерів).
func (c *Customer) GetReceived() []Message {
	return c.received
}

// Manager — менеджер підтримки, учасник чату.
// Зберігає посилання тільки на медіатора, не на інших учасників.
type Manager struct {
	name     string
	mediator ChatMediator
	received []Message
}

// NewManager — конструктор Manager.
func NewManager(name string, mediator ChatMediator) *Manager {
	return &Manager{
		name:     name,
		mediator: mediator,
		received: []Message{},
	}
}

// GetName — повертає ім'я менеджера.
func (g *Manager) GetName() string {
	return g.name
}

// Receive — отримує повідомлення від іншого учасника через медіатора.
func (g *Manager) Receive(from, message string) {
	g.received = append(g.received, Message{From: from, To: g.name, Content: message})
}

// GetReceived — повертає всі отримані повідомлення (для тестів та хендлерів).
func (g *Manager) GetReceived() []Message {
	return g.received
}
