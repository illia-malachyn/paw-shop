package discount

import (
	"testing"

	"github.com/illia-malachyn/paw-shop/internal/notification"
)

func TestApplyDiscountCommand_Execute(t *testing.T) {
	subject := notification.NewPriceSubject()
	subject.SetPrice("p1", 1000)

	cmd := &ApplyDiscountCommand{
		ProductID: "p1",
		Strategy:  &PercentStrategy{Percent: 10},
		Subject:   subject,
	}

	newPrice := cmd.Execute()
	if !almostEqual(newPrice, 900) {
		t.Errorf("expected 900, got %.2f", newPrice)
	}

	price, _ := subject.GetPrice("p1")
	if !almostEqual(price, 900) {
		t.Errorf("expected subject price 900, got %.2f", price)
	}
}

func TestApplyDiscountCommand_Undo(t *testing.T) {
	subject := notification.NewPriceSubject()
	subject.SetPrice("p1", 1000)

	cmd := &ApplyDiscountCommand{
		ProductID: "p1",
		Strategy:  &FixedStrategy{Amount: 200},
		Subject:   subject,
	}

	cmd.Execute()
	restored := cmd.Undo()

	if !almostEqual(restored, 1000) {
		t.Errorf("expected restored price 1000, got %.2f", restored)
	}

	price, _ := subject.GetPrice("p1")
	if !almostEqual(price, 1000) {
		t.Errorf("expected subject price 1000, got %.2f", price)
	}
}

func TestCommandHistory(t *testing.T) {
	subject := notification.NewPriceSubject()
	subject.SetPrice("p1", 1000)

	history := NewCommandHistory()

	// Застосовуємо дві знижки послідовно
	cmd1 := &ApplyDiscountCommand{
		ProductID: "p1",
		Strategy:  &PercentStrategy{Percent: 10},
		Subject:   subject,
	}
	history.Execute(cmd1) // 1000 -> 900

	cmd2 := &ApplyDiscountCommand{
		ProductID: "p1",
		Strategy:  &FixedStrategy{Amount: 100},
		Subject:   subject,
	}
	history.Execute(cmd2) // 900 -> 800

	price, _ := subject.GetPrice("p1")
	if !almostEqual(price, 800) {
		t.Errorf("expected 800 after two discounts, got %.2f", price)
	}

	// Undo останню
	history.Undo() // 800 -> 900
	price, _ = subject.GetPrice("p1")
	if !almostEqual(price, 900) {
		t.Errorf("expected 900 after first undo, got %.2f", price)
	}

	// Undo першу
	history.Undo() // 900 -> 1000
	price, _ = subject.GetPrice("p1")
	if !almostEqual(price, 1000) {
		t.Errorf("expected 1000 after second undo, got %.2f", price)
	}

	if history.HasHistory() {
		t.Error("expected empty history after all undos")
	}
}
