package order

import (
	"testing"
)

func TestConfirmOrderCommand(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus string
		wantStatus    string
		wantErr       bool
	}{
		{
			name:          "confirm new order succeeds",
			initialStatus: "new",
			wantStatus:    "confirmed",
			wantErr:       false,
		},
		{
			name:          "confirm already-confirmed order returns error",
			initialStatus: "confirmed",
			wantStatus:    "confirmed",
			wantErr:       true,
		},
		{
			name:          "confirm rejected order returns error",
			initialStatus: "rejected",
			wantStatus:    "rejected",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Order{ID: "order-1", Status: tt.initialStatus, Items: []string{"kibble"}}
			cmd := &ConfirmOrderCommand{Order: o}

			err := cmd.Execute()

			if (err != nil) != tt.wantErr {
				t.Errorf("ConfirmOrderCommand.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
			if o.Status != tt.wantStatus {
				t.Errorf("order status = %q, want %q", o.Status, tt.wantStatus)
			}
		})
	}
}

func TestRejectOrderCommand(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus string
		wantStatus    string
		wantErr       bool
	}{
		{
			name:          "reject new order succeeds",
			initialStatus: "new",
			wantStatus:    "rejected",
			wantErr:       false,
		},
		{
			name:          "reject already-confirmed order returns error",
			initialStatus: "confirmed",
			wantStatus:    "confirmed",
			wantErr:       true,
		},
		{
			name:          "reject already-rejected order returns error",
			initialStatus: "rejected",
			wantStatus:    "rejected",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Order{ID: "order-2", Status: tt.initialStatus, Items: []string{"wet-food"}}
			cmd := &RejectOrderCommand{Order: o}

			err := cmd.Execute()

			if (err != nil) != tt.wantErr {
				t.Errorf("RejectOrderCommand.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
			if o.Status != tt.wantStatus {
				t.Errorf("order status = %q, want %q", o.Status, tt.wantStatus)
			}
		})
	}
}

func TestMacroCommand(t *testing.T) {
	t.Run("executes multiple commands sequentially", func(t *testing.T) {
		o1 := &Order{ID: "o1", Status: "new", Items: []string{"item1"}}
		o2 := &Order{ID: "o2", Status: "new", Items: []string{"item2"}}
		o3 := &Order{ID: "o3", Status: "new", Items: []string{"item3"}}

		macro := NewMacroCommand([]OrderCommand{
			&ConfirmOrderCommand{Order: o1},
			&ConfirmOrderCommand{Order: o2},
			&ConfirmOrderCommand{Order: o3},
		})

		err := macro.Execute()

		if err != nil {
			t.Errorf("MacroCommand.Execute() unexpected error = %v", err)
		}
		if o1.Status != "confirmed" {
			t.Errorf("o1 status = %q, want \"confirmed\"", o1.Status)
		}
		if o2.Status != "confirmed" {
			t.Errorf("o2 status = %q, want \"confirmed\"", o2.Status)
		}
		if o3.Status != "confirmed" {
			t.Errorf("o3 status = %q, want \"confirmed\"", o3.Status)
		}
	})

	t.Run("stops on first error", func(t *testing.T) {
		o1 := &Order{ID: "o1", Status: "new", Items: []string{"item1"}}
		o2 := &Order{ID: "o2", Status: "confirmed", Items: []string{"item2"}} // already confirmed
		o3 := &Order{ID: "o3", Status: "new", Items: []string{"item3"}}

		macro := NewMacroCommand([]OrderCommand{
			&ConfirmOrderCommand{Order: o1},
			&ConfirmOrderCommand{Order: o2},
			&ConfirmOrderCommand{Order: o3},
		})

		err := macro.Execute()

		if err == nil {
			t.Error("MacroCommand.Execute() expected error, got nil")
		}
		// o1 should be confirmed, o3 should remain new (stopped before it)
		if o1.Status != "confirmed" {
			t.Errorf("o1 status = %q, want \"confirmed\"", o1.Status)
		}
		if o3.Status != "new" {
			t.Errorf("o3 status = %q, want \"new\" (should not have been processed)", o3.Status)
		}
	})

	t.Run("empty commands slice returns nil", func(t *testing.T) {
		macro := NewMacroCommand([]OrderCommand{})

		err := macro.Execute()

		if err != nil {
			t.Errorf("MacroCommand.Execute() with empty commands returned error = %v, want nil", err)
		}
	})
}
