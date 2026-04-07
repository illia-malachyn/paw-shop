package notification

import "testing"

func TestPriceSubject_SetAndGet(t *testing.T) {
	s := NewPriceSubject()
	s.SetPrice("p1", 100)

	price, ok := s.GetPrice("p1")
	if !ok || price != 100 {
		t.Errorf("expected price 100, got %.2f (ok=%v)", price, ok)
	}

	_, ok = s.GetPrice("nonexistent")
	if ok {
		t.Error("expected ok=false for nonexistent product")
	}
}

func TestPriceSubject_NotifiesObservers(t *testing.T) {
	s := NewPriceSubject()
	s.SetPrice("p1", 100)

	obs := &InMemoryObserver{UserEmail: "test@test.com"}
	s.Subscribe("p1", obs)

	// Зміна ціни — має прийти сповіщення
	s.SetPrice("p1", 80)

	if len(obs.Records) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(obs.Records))
	}

	r := obs.Records[0]
	if r.OldPrice != 100 || r.NewPrice != 80 {
		t.Errorf("expected 100->80, got %.2f->%.2f", r.OldPrice, r.NewPrice)
	}
	if r.ProductID != "p1" {
		t.Errorf("expected product_id 'p1', got '%s'", r.ProductID)
	}
}

func TestPriceSubject_NoNotificationOnSamePrice(t *testing.T) {
	s := NewPriceSubject()
	s.SetPrice("p1", 100)

	obs := &InMemoryObserver{UserEmail: "test@test.com"}
	s.Subscribe("p1", obs)

	// Та сама ціна — сповіщення не має бути
	s.SetPrice("p1", 100)

	if len(obs.Records) != 0 {
		t.Errorf("expected 0 notifications for same price, got %d", len(obs.Records))
	}
}

func TestPriceSubject_MultipleObservers(t *testing.T) {
	s := NewPriceSubject()
	s.SetPrice("p1", 500)

	obs1 := &InMemoryObserver{UserEmail: "user1@test.com"}
	obs2 := &InMemoryObserver{UserEmail: "user2@test.com"}
	s.Subscribe("p1", obs1)
	s.Subscribe("p1", obs2)

	s.SetPrice("p1", 400)

	if len(obs1.Records) != 1 {
		t.Errorf("obs1: expected 1 notification, got %d", len(obs1.Records))
	}
	if len(obs2.Records) != 1 {
		t.Errorf("obs2: expected 1 notification, got %d", len(obs2.Records))
	}
}

func TestPriceSubject_DifferentProducts(t *testing.T) {
	s := NewPriceSubject()
	s.SetPrice("p1", 100)
	s.SetPrice("p2", 200)

	obs := &InMemoryObserver{UserEmail: "test@test.com"}
	s.Subscribe("p1", obs) // підписка тільки на p1

	s.SetPrice("p2", 150) // зміна p2 — не має нотифікувати

	if len(obs.Records) != 0 {
		t.Errorf("expected 0 notifications for unsubscribed product, got %d", len(obs.Records))
	}
}
