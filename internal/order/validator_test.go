package order

import (
	"strings"
	"testing"
)

func TestValidationChain(t *testing.T) {
	tests := []struct {
		name        string
		req         *OrderRequest
		wantErr     bool
		errContains string
	}{
		{
			name: "valid request passes full chain",
			req: &OrderRequest{
				Items:   []string{"kibble"},
				Address: "123 Dog St",
				Amount:  100.0,
			},
			wantErr: false,
		},
		{
			name: "stock failure — out-of-stock item",
			req: &OrderRequest{
				Items:   []string{"kibble", "out-of-stock-item"},
				Address: "123 Dog St",
				Amount:  100.0,
			},
			wantErr:     true,
			errContains: "stock validation failed",
		},
		{
			name: "address failure — empty string",
			req: &OrderRequest{
				Items:   []string{"kibble"},
				Address: "",
				Amount:  100.0,
			},
			wantErr:     true,
			errContains: "address validation failed",
		},
		{
			name: "address failure — whitespace only",
			req: &OrderRequest{
				Items:   []string{"kibble"},
				Address: "   ",
				Amount:  100.0,
			},
			wantErr:     true,
			errContains: "address validation failed",
		},
		{
			name: "payment failure — amount is zero",
			req: &OrderRequest{
				Items:   []string{"kibble"},
				Address: "123 Dog St",
				Amount:  0,
			},
			wantErr:     true,
			errContains: "payment validation failed",
		},
		{
			name: "payment failure — negative amount",
			req: &OrderRequest{
				Items:   []string{"kibble"},
				Address: "123 Dog St",
				Amount:  -10,
			},
			wantErr:     true,
			errContains: "payment validation failed",
		},
		{
			name: "chain ordering — stock error takes priority over address error",
			req: &OrderRequest{
				Items:   []string{"out-of-stock-item"},
				Address: "",
				Amount:  100.0,
			},
			wantErr:     true,
			errContains: "stock validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chain := NewValidationChain()

			err := chain.Validate(tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("Validate() error = %q, want error containing %q", err.Error(), tt.errContains)
			}
		})
	}
}

func TestSingleValidator(t *testing.T) {
	t.Run("StockValidator without SetNext validates independently", func(t *testing.T) {
		v := &StockValidator{}

		// valid items — should pass (no next, returns nil)
		err := v.Validate(&OrderRequest{
			Items:   []string{"kibble"},
			Address: "",  // empty address is fine — AddressValidator not in chain
			Amount:  0,   // zero amount is fine — PaymentValidator not in chain
		})
		if err != nil {
			t.Errorf("StockValidator.Validate() with valid items unexpected error = %v", err)
		}

		// out-of-stock item — should fail
		err = v.Validate(&OrderRequest{
			Items:   []string{"out-of-stock-item"},
			Address: "123 Dog St",
			Amount:  100.0,
		})
		if err == nil {
			t.Error("StockValidator.Validate() with out-of-stock item expected error, got nil")
		}
		if !strings.Contains(err.Error(), "stock validation failed") {
			t.Errorf("StockValidator.Validate() error = %q, want containing \"stock validation failed\"", err.Error())
		}
	})
}
