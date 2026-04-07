package order

import (
	"strings"
	"testing"
)

func TestStateTransitions(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *Order
		action      func(o *Order) error
		wantState   string
		wantErr     bool
		errContains string
	}{
		{
			name:      "NewState.Next transitions to confirmed",
			setup:     func() *Order { return NewOrder("o1", nil) },
			action:    func(o *Order) error { return o.Next() },
			wantState: "confirmed",
			wantErr:   false,
		},
		{
			name:      "NewState.Cancel transitions to cancelled",
			setup:     func() *Order { return NewOrder("o2", nil) },
			action:    func(o *Order) error { return o.Cancel() },
			wantState: "cancelled",
			wantErr:   false,
		},
		{
			name: "ConfirmedState.Next transitions to shipped",
			setup: func() *Order {
				o := NewOrder("o3", nil)
				_ = o.Next() // new -> confirmed
				return o
			},
			action:    func(o *Order) error { return o.Next() },
			wantState: "shipped",
			wantErr:   false,
		},
		{
			name: "ConfirmedState.Cancel transitions to cancelled",
			setup: func() *Order {
				o := NewOrder("o4", nil)
				_ = o.Next() // new -> confirmed
				return o
			},
			action:    func(o *Order) error { return o.Cancel() },
			wantState: "cancelled",
			wantErr:   false,
		},
		{
			name: "ShippedState.Next transitions to delivered",
			setup: func() *Order {
				o := NewOrder("o5", nil)
				_ = o.Next() // new -> confirmed
				_ = o.Next() // confirmed -> shipped
				return o
			},
			action:    func(o *Order) error { return o.Next() },
			wantState: "delivered",
			wantErr:   false,
		},
		{
			name: "ShippedState.Cancel returns error",
			setup: func() *Order {
				o := NewOrder("o6", nil)
				_ = o.Next() // new -> confirmed
				_ = o.Next() // confirmed -> shipped
				return o
			},
			action:      func(o *Order) error { return o.Cancel() },
			wantState:   "shipped",
			wantErr:     true,
			errContains: "cannot cancel order: already shipped",
		},
		{
			name: "DeliveredState.Next returns error",
			setup: func() *Order {
				o := NewOrder("o7", nil)
				_ = o.Next() // new -> confirmed
				_ = o.Next() // confirmed -> shipped
				_ = o.Next() // shipped -> delivered
				return o
			},
			action:      func(o *Order) error { return o.Next() },
			wantState:   "delivered",
			wantErr:     true,
			errContains: "order already delivered",
		},
		{
			name: "DeliveredState.Cancel returns error",
			setup: func() *Order {
				o := NewOrder("o8", nil)
				_ = o.Next() // new -> confirmed
				_ = o.Next() // confirmed -> shipped
				_ = o.Next() // shipped -> delivered
				return o
			},
			action:      func(o *Order) error { return o.Cancel() },
			wantState:   "delivered",
			wantErr:     true,
			errContains: "cannot cancel order: already delivered",
		},
		{
			name: "CancelledState.Next returns error",
			setup: func() *Order {
				o := NewOrder("o9", nil)
				_ = o.Cancel() // new -> cancelled
				return o
			},
			action:      func(o *Order) error { return o.Next() },
			wantState:   "cancelled",
			wantErr:     true,
			errContains: "order is cancelled",
		},
		{
			name: "CancelledState.Cancel returns error",
			setup: func() *Order {
				o := NewOrder("o10", nil)
				_ = o.Cancel() // new -> cancelled
				return o
			},
			action:      func(o *Order) error { return o.Cancel() },
			wantState:   "cancelled",
			wantErr:     true,
			errContains: "order is already cancelled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.setup()

			err := tt.action(o)

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error message = %q, want to contain %q", err.Error(), tt.errContains)
				}
			}
			if o.GetState().Name() != tt.wantState {
				t.Errorf("state = %q, want %q", o.GetState().Name(), tt.wantState)
			}
			if o.Status != tt.wantState {
				t.Errorf("Status field = %q, want %q (should be in sync with state)", o.Status, tt.wantState)
			}
		})
	}
}

func TestFullLifecycle(t *testing.T) {
	o := NewOrder("lifecycle-1", []string{"kibble"})

	if o.GetState().Name() != "new" {
		t.Fatalf("initial state = %q, want \"new\"", o.GetState().Name())
	}

	if err := o.Next(); err != nil {
		t.Fatalf("Next() from new: unexpected error %v", err)
	}
	if o.GetState().Name() != "confirmed" {
		t.Errorf("after 1st Next: state = %q, want \"confirmed\"", o.GetState().Name())
	}

	if err := o.Next(); err != nil {
		t.Fatalf("Next() from confirmed: unexpected error %v", err)
	}
	if o.GetState().Name() != "shipped" {
		t.Errorf("after 2nd Next: state = %q, want \"shipped\"", o.GetState().Name())
	}

	if err := o.Next(); err != nil {
		t.Fatalf("Next() from shipped: unexpected error %v", err)
	}
	if o.GetState().Name() != "delivered" {
		t.Errorf("after 3rd Next: state = %q, want \"delivered\"", o.GetState().Name())
	}
}
