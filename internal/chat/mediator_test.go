package chat

import (
	"testing"
)

// TestSendMessage — клієнт надсилає повідомлення менеджеру, менеджер отримує його.
func TestSendMessage(t *testing.T) {
	m := NewSupportChatMediator()

	customer := NewCustomer("customer1", m)
	manager := NewManager("manager1", m)

	m.AddParticipant(customer)
	m.AddParticipant(manager)

	m.SendMessage("customer1", "manager1", "Hello!")

	got := manager.GetReceived()
	if len(got) != 1 {
		t.Fatalf("manager expected 1 message, got %d", len(got))
	}
	if got[0].From != "customer1" {
		t.Errorf("expected From=customer1, got %s", got[0].From)
	}
	if got[0].Content != "Hello!" {
		t.Errorf("expected Content=Hello!, got %s", got[0].Content)
	}

	if len(customer.GetReceived()) != 0 {
		t.Errorf("customer should have 0 received messages, got %d", len(customer.GetReceived()))
	}
}

// TestBidirectionalRouting — двостороннє спілкування: від клієнта до менеджера та навпаки.
func TestBidirectionalRouting(t *testing.T) {
	m := NewSupportChatMediator()

	customer := NewCustomer("customer1", m)
	manager := NewManager("manager1", m)

	m.AddParticipant(customer)
	m.AddParticipant(manager)

	m.SendMessage("customer1", "manager1", "Hi manager!")
	m.SendMessage("manager1", "customer1", "Hi customer!")

	customerReceived := customer.GetReceived()
	if len(customerReceived) != 1 {
		t.Fatalf("customer expected 1 message, got %d", len(customerReceived))
	}
	if customerReceived[0].From != "manager1" {
		t.Errorf("customer expected message from manager1, got %s", customerReceived[0].From)
	}

	managerReceived := manager.GetReceived()
	if len(managerReceived) != 1 {
		t.Fatalf("manager expected 1 message, got %d", len(managerReceived))
	}
	if managerReceived[0].From != "customer1" {
		t.Errorf("manager expected message from customer1, got %s", managerReceived[0].From)
	}
}

// TestParticipantIsolation — учасники не мають прямих посилань один на одного;
// маршрутизація відбувається виключно через медіатора.
func TestParticipantIsolation(t *testing.T) {
	m := NewSupportChatMediator()

	customer := NewCustomer("customer1", m)
	manager := NewManager("manager1", m)

	m.AddParticipant(customer)
	m.AddParticipant(manager)

	// Відправляємо повідомлення — якщо маршрутизація не через медіатора, менеджер не отримає
	m.SendMessage("customer1", "manager1", "test isolation")

	if len(manager.GetReceived()) != 1 {
		t.Errorf("message should be routed via mediator; manager expected 1 message")
	}

	// Видаляємо менеджера з медіатора (симулюємо відсутність прямого посилання)
	m2 := NewSupportChatMediator()
	m2.AddParticipant(customer)
	// manager не зареєстрований у m2

	m2.SendMessage("customer1", "manager1", "no route")
	// Менеджер не повинен отримати це повідомлення (посилання немає у медіатора)
	if len(manager.GetReceived()) != 1 {
		t.Errorf("unregistered manager should not receive message via different mediator")
	}
}

// TestGetHistory — перевіряє фільтрацію історії за учасником.
func TestGetHistory(t *testing.T) {
	m := NewSupportChatMediator()

	customer := NewCustomer("customer1", m)
	manager := NewManager("manager1", m)

	m.AddParticipant(customer)
	m.AddParticipant(manager)

	m.SendMessage("customer1", "manager1", "msg1")
	m.SendMessage("manager1", "customer1", "msg2")
	m.SendMessage("customer1", "manager1", "msg3")

	customerHistory := m.GetHistory("customer1")
	if len(customerHistory) != 3 {
		t.Errorf("customer1 history expected 3 messages, got %d", len(customerHistory))
	}

	managerHistory := m.GetHistory("manager1")
	if len(managerHistory) != 3 {
		t.Errorf("manager1 history expected 3 messages, got %d", len(managerHistory))
	}

	unknownHistory := m.GetHistory("unknown")
	if len(unknownHistory) != 0 {
		t.Errorf("unknown participant history expected 0 messages, got %d", len(unknownHistory))
	}
}

// TestUnknownRecipient — надсилання до незареєстрованого учасника не паніку не викликає,
// але повідомлення зберігається в історії.
func TestUnknownRecipient(t *testing.T) {
	m := NewSupportChatMediator()

	customer := NewCustomer("customer1", m)
	m.AddParticipant(customer)

	// Не повинно виникати паніки
	m.SendMessage("customer1", "nobody", "test")

	history := m.GetHistory("customer1")
	if len(history) != 1 {
		t.Errorf("message to unknown recipient should still be stored; expected 1 in history, got %d", len(history))
	}
}

// TestMultipleParticipants — кілька клієнтів та один менеджер; менеджер отримує від кожного.
func TestMultipleParticipants(t *testing.T) {
	m := NewSupportChatMediator()

	customer1 := NewCustomer("customer1", m)
	customer2 := NewCustomer("customer2", m)
	manager := NewManager("manager1", m)

	m.AddParticipant(customer1)
	m.AddParticipant(customer2)
	m.AddParticipant(manager)

	m.SendMessage("customer1", "manager1", "message from customer1")
	m.SendMessage("customer2", "manager1", "message from customer2")

	received := manager.GetReceived()
	if len(received) != 2 {
		t.Fatalf("manager expected 2 messages, got %d", len(received))
	}

	senders := map[string]bool{}
	for _, msg := range received {
		senders[msg.From] = true
	}

	if !senders["customer1"] {
		t.Errorf("manager expected message from customer1")
	}
	if !senders["customer2"] {
		t.Errorf("manager expected message from customer2")
	}
}
