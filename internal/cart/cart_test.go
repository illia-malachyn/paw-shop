package cart

import "testing"

func TestCartAddItem(t *testing.T) {
	c := &Cart{}

	c.AddItem(CartItem{ProductID: "p1", Name: "Food", Price: 100, Quantity: 1})

	if len(c.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(c.Items))
	}
	if c.Items[0].ProductID != "p1" {
		t.Errorf("expected product_id p1, got %s", c.Items[0].ProductID)
	}

	// Adding same ProductID increments quantity
	c.AddItem(CartItem{ProductID: "p1", Name: "Food", Price: 100, Quantity: 2})

	if len(c.Items) != 1 {
		t.Fatalf("expected 1 item after re-add, got %d", len(c.Items))
	}
	if c.Items[0].Quantity != 3 {
		t.Errorf("expected quantity 3, got %d", c.Items[0].Quantity)
	}
}

func TestCartRemoveItem(t *testing.T) {
	c := &Cart{}
	c.AddItem(CartItem{ProductID: "p1", Name: "Food", Price: 100, Quantity: 1})

	err := c.RemoveItem("p1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(c.Items) != 0 {
		t.Errorf("expected 0 items after remove, got %d", len(c.Items))
	}

	err = c.RemoveItem("nonexistent")
	if err == nil {
		t.Error("expected error when removing non-existent item, got nil")
	}
}

func TestCartUpdateQuantity(t *testing.T) {
	c := &Cart{}
	c.AddItem(CartItem{ProductID: "p1", Name: "Food", Price: 100, Quantity: 1})

	err := c.UpdateQuantity("p1", 5)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if c.Items[0].Quantity != 5 {
		t.Errorf("expected quantity 5, got %d", c.Items[0].Quantity)
	}

	// Update to 0 removes item
	err = c.UpdateQuantity("p1", 0)
	if err != nil {
		t.Fatalf("expected no error on zero quantity, got %v", err)
	}
	if len(c.Items) != 0 {
		t.Errorf("expected 0 items after quantity set to 0, got %d", len(c.Items))
	}

	// Update non-existent returns error
	err = c.UpdateQuantity("nonexistent", 3)
	if err == nil {
		t.Error("expected error for non-existent product, got nil")
	}
}

func TestCartSaveRestore(t *testing.T) {
	c := &Cart{}
	c.AddItem(CartItem{ProductID: "p1", Name: "Food", Price: 100, Quantity: 1})
	c.AddItem(CartItem{ProductID: "p2", Name: "Treat", Price: 50, Quantity: 2})

	memento := c.Save()

	c.AddItem(CartItem{ProductID: "p3", Name: "WetFood", Price: 75, Quantity: 1})
	if len(c.Items) != 3 {
		t.Fatalf("expected 3 items after adding third, got %d", len(c.Items))
	}

	c.Restore(memento)

	if len(c.Items) != 2 {
		t.Fatalf("expected 2 items after restore, got %d", len(c.Items))
	}
	if c.Items[0].ProductID != "p1" || c.Items[1].ProductID != "p2" {
		t.Errorf("restored items do not match saved state")
	}
}

func TestMementoDeepCopy(t *testing.T) {
	c := &Cart{}
	c.AddItem(CartItem{ProductID: "p1", Name: "Food", Price: 100, Quantity: 1})

	memento := c.Save()

	// Mutate cart after saving
	c.AddItem(CartItem{ProductID: "p2", Name: "Treat", Price: 50, Quantity: 1})
	if len(c.Items) != 2 {
		t.Fatalf("expected 2 items after add, got %d", len(c.Items))
	}

	// Restore from memento — should have only original item
	c.Restore(memento)

	if len(c.Items) != 1 {
		t.Fatalf("expected 1 item after restore, got %d — shallow copy bug detected", len(c.Items))
	}
	if c.Items[0].ProductID != "p1" {
		t.Errorf("expected p1 after restore, got %s", c.Items[0].ProductID)
	}
}

func TestCartHistoryPushPop(t *testing.T) {
	c := &Cart{}
	h := &CartHistory{}

	c.AddItem(CartItem{ProductID: "p1", Name: "Food", Price: 100, Quantity: 1})
	m1 := c.Save()
	h.Push(m1)

	c.AddItem(CartItem{ProductID: "p2", Name: "Treat", Price: 50, Quantity: 1})
	m2 := c.Save()
	h.Push(m2)

	c.AddItem(CartItem{ProductID: "p3", Name: "WetFood", Price: 75, Quantity: 1})
	m3 := c.Save()
	h.Push(m3)

	// Pop returns LIFO order
	got, ok := h.Pop()
	if !ok {
		t.Fatal("expected ok=true on first pop")
	}
	if len(got.items) != 3 {
		t.Errorf("expected 3 items in third memento, got %d", len(got.items))
	}

	got, ok = h.Pop()
	if !ok {
		t.Fatal("expected ok=true on second pop")
	}
	if len(got.items) != 2 {
		t.Errorf("expected 2 items in second memento, got %d", len(got.items))
	}

	got, ok = h.Pop()
	if !ok {
		t.Fatal("expected ok=true on third pop")
	}
	if len(got.items) != 1 {
		t.Errorf("expected 1 item in first memento, got %d", len(got.items))
	}

	// Pop from empty returns false
	_, ok = h.Pop()
	if ok {
		t.Error("expected ok=false on pop from empty stack")
	}
}

func TestUndoAfterAdd(t *testing.T) {
	c := &Cart{}
	h := &CartHistory{}

	// Save empty state
	h.Push(c.Save())

	// Add item
	c.AddItem(CartItem{ProductID: "p1", Name: "Food", Price: 100, Quantity: 1})
	if len(c.Items) != 1 {
		t.Fatalf("expected 1 item after add, got %d", len(c.Items))
	}

	// Undo by restoring from history
	m, ok := h.Pop()
	if !ok {
		t.Fatal("expected history to have entry")
	}
	c.Restore(m)

	if len(c.Items) != 0 {
		t.Errorf("expected 0 items after undo, got %d", len(c.Items))
	}
}

func TestUndoAfterRemove(t *testing.T) {
	c := &Cart{}
	h := &CartHistory{}

	c.AddItem(CartItem{ProductID: "p1", Name: "Food", Price: 100, Quantity: 1})
	c.AddItem(CartItem{ProductID: "p2", Name: "Treat", Price: 50, Quantity: 2})

	// Save before remove
	h.Push(c.Save())

	// Remove one item
	err := c.RemoveItem("p1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.Items) != 1 {
		t.Fatalf("expected 1 item after remove, got %d", len(c.Items))
	}

	// Undo — should restore both items
	m, ok := h.Pop()
	if !ok {
		t.Fatal("expected history to have entry")
	}
	c.Restore(m)

	if len(c.Items) != 2 {
		t.Errorf("expected 2 items after undo, got %d", len(c.Items))
	}
}
