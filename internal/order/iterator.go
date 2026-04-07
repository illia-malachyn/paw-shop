package order

// OrderIterator — інтерфейс ітератора замовлень (Iterator pattern).
// Забезпечує послідовний доступ до елементів колекції без розкриття її внутрішньої структури.
type OrderIterator interface {
	HasNext() bool
	Next() *Order
}

// OrderCollection — колекція замовлень з підтримкою ітерації.
type OrderCollection struct {
	orders []*Order
}

// NewOrderCollection — створює порожню колекцію замовлень.
func NewOrderCollection() *OrderCollection {
	return &OrderCollection{}
}

// Add — додає замовлення до колекції.
func (c *OrderCollection) Add(order *Order) {
	c.orders = append(c.orders, order)
}

// CreateIterator — повертає ітератор, що обходить усі замовлення колекції.
func (c *OrderCollection) CreateIterator() OrderIterator {
	return &allIterator{orders: c.orders, index: 0}
}

// CreateFilteredIterator — повертає ітератор, що повертає тільки замовлення із заданим статусом.
func (c *OrderCollection) CreateFilteredIterator(status string) OrderIterator {
	return &filteredIterator{orders: c.orders, status: status, index: 0}
}

// GetByID — шукає замовлення за ідентифікатором. Повертає false, якщо не знайдено.
func (c *OrderCollection) GetByID(id string) (*Order, bool) {
	for _, o := range c.orders {
		if o.ID == id {
			return o, true
		}
	}
	return nil, false
}

// Count — повертає кількість замовлень у колекції.
func (c *OrderCollection) Count() int {
	return len(c.orders)
}

// allIterator — ітератор, що обходить усі замовлення підряд.
type allIterator struct {
	orders []*Order
	index  int
}

// HasNext — повертає true, якщо є ще непереглянуті замовлення.
func (it *allIterator) HasNext() bool {
	return it.index < len(it.orders)
}

// Next — повертає поточне замовлення і переміщує курсор вперед.
func (it *allIterator) Next() *Order {
	if !it.HasNext() {
		return nil
	}
	o := it.orders[it.index]
	it.index++
	return o
}

// filteredIterator — ітератор, що повертає тільки замовлення із заданим статусом.
type filteredIterator struct {
	orders []*Order
	status string
	index  int
}

// HasNext — шукає наступне замовлення з відповідним статусом починаючи з поточної позиції.
func (it *filteredIterator) HasNext() bool {
	for it.index < len(it.orders) {
		if it.orders[it.index].GetState().Name() == it.status {
			return true
		}
		it.index++
	}
	return false
}

// Next — повертає поточне замовлення з відповідним статусом і переміщує курсор вперед.
func (it *filteredIterator) Next() *Order {
	if !it.HasNext() {
		return nil
	}
	o := it.orders[it.index]
	it.index++
	return o
}
