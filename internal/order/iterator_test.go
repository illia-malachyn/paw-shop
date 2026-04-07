package order

import "testing"

func TestCreateIterator(t *testing.T) {
	t.Run("iterates all orders", func(t *testing.T) {
		c := NewOrderCollection()
		c.Add(NewOrder("a", nil))
		c.Add(NewOrder("b", nil))
		c.Add(NewOrder("c", nil))

		it := c.CreateIterator()
		count := 0
		for it.HasNext() {
			o := it.Next()
			if o == nil {
				t.Fatal("Next() returned nil when HasNext() was true")
			}
			count++
		}
		if count != 3 {
			t.Errorf("iterated %d orders, want 3", count)
		}
	})

	t.Run("empty collection returns HasNext false immediately", func(t *testing.T) {
		c := NewOrderCollection()
		it := c.CreateIterator()
		if it.HasNext() {
			t.Error("HasNext() = true for empty collection, want false")
		}
		if it.Next() != nil {
			t.Error("Next() on empty iterator should return nil")
		}
	})
}

func TestCreateFilteredIterator(t *testing.T) {
	t.Run("returns only confirmed orders", func(t *testing.T) {
		c := NewOrderCollection()

		o1 := NewOrder("o1", nil) // new
		o2 := NewOrder("o2", nil)
		_ = o2.Next() // confirmed
		o3 := NewOrder("o3", nil)
		_ = o3.Next() // confirmed
		_ = o3.Next() // shipped

		c.Add(o1)
		c.Add(o2)
		c.Add(o3)

		it := c.CreateFilteredIterator("confirmed")
		var results []*Order
		for it.HasNext() {
			results = append(results, it.Next())
		}

		if len(results) != 1 {
			t.Fatalf("filtered iterator returned %d orders, want 1", len(results))
		}
		if results[0].ID != "o2" {
			t.Errorf("filtered result ID = %q, want \"o2\"", results[0].ID)
		}
	})

	t.Run("filtered iterator with no matches returns HasNext false", func(t *testing.T) {
		c := NewOrderCollection()
		c.Add(NewOrder("a", nil))
		c.Add(NewOrder("b", nil))

		it := c.CreateFilteredIterator("delivered")
		if it.HasNext() {
			t.Error("HasNext() = true, want false when no matching orders")
		}
	})

	t.Run("multiple matching statuses returned in order", func(t *testing.T) {
		c := NewOrderCollection()
		o1 := NewOrder("x1", nil)
		_ = o1.Next() // confirmed
		o2 := NewOrder("x2", nil)
		_ = o2.Cancel() // cancelled
		o3 := NewOrder("x3", nil)
		_ = o3.Next() // confirmed

		c.Add(o1)
		c.Add(o2)
		c.Add(o3)

		it := c.CreateFilteredIterator("confirmed")
		var ids []string
		for it.HasNext() {
			ids = append(ids, it.Next().ID)
		}

		if len(ids) != 2 {
			t.Fatalf("got %d results, want 2", len(ids))
		}
		if ids[0] != "x1" || ids[1] != "x3" {
			t.Errorf("order IDs = %v, want [x1 x3]", ids)
		}
	})
}

func TestGetByID(t *testing.T) {
	c := NewOrderCollection()
	c.Add(NewOrder("alpha", []string{"item1"}))
	c.Add(NewOrder("beta", []string{"item2"}))

	t.Run("returns correct order when found", func(t *testing.T) {
		o, ok := c.GetByID("alpha")
		if !ok {
			t.Fatal("GetByID(\"alpha\") = false, want true")
		}
		if o.ID != "alpha" {
			t.Errorf("returned order ID = %q, want \"alpha\"", o.ID)
		}
	})

	t.Run("returns false for nonexistent ID", func(t *testing.T) {
		_, ok := c.GetByID("nonexistent")
		if ok {
			t.Error("GetByID(\"nonexistent\") = true, want false")
		}
	})
}

func TestOrderCollectionCount(t *testing.T) {
	c := NewOrderCollection()
	if c.Count() != 0 {
		t.Errorf("Count() = %d, want 0 for empty collection", c.Count())
	}

	c.Add(NewOrder("1", nil))
	c.Add(NewOrder("2", nil))

	if c.Count() != 2 {
		t.Errorf("Count() = %d, want 2", c.Count())
	}
}
