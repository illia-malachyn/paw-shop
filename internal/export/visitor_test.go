package export

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/illia-malachyn/paw-shop/internal/cart"
)

// jsonExportResult — допоміжна структура для розбору JSON-виводу JSONExportVisitor.
type jsonExportResult struct {
	Items []struct {
		ProductID string  `json:"product_id"`
		Name      string  `json:"name"`
		Price     float64 `json:"price"`
		Quantity  int     `json:"quantity"`
		Subtotal  float64 `json:"subtotal"`
	} `json:"items"`
	Total float64 `json:"total"`
}

func TestJSONExportSingleItem(t *testing.T) {
	c := cart.Cart{
		Items: []cart.CartItem{
			{ProductID: "p1", Name: "Royal Canin Dry", Price: 100, Quantity: 2},
		},
	}

	v := &JSONExportVisitor{}
	result := ExportCart(c, v)

	var parsed jsonExportResult
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("Result is not valid JSON: %v\nGot: %s", err, result)
	}

	if len(parsed.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(parsed.Items))
	}

	if parsed.Items[0].Subtotal != 200 {
		t.Errorf("expected subtotal 200, got %.2f", parsed.Items[0].Subtotal)
	}

	if parsed.Total != 200 {
		t.Errorf("expected total 200, got %.2f", parsed.Total)
	}
}

func TestJSONExportMultipleItems(t *testing.T) {
	c := cart.Cart{
		Items: []cart.CartItem{
			{ProductID: "p1", Name: "Royal Canin Dry", Price: 100, Quantity: 2},
			{ProductID: "p2", Name: "Acana Adult", Price: 150, Quantity: 1},
			{ProductID: "p3", Name: "Orijen Puppy", Price: 200, Quantity: 3},
		},
	}

	v := &JSONExportVisitor{}
	result := ExportCart(c, v)

	var parsed jsonExportResult
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("Result is not valid JSON: %v\nGot: %s", err, result)
	}

	if len(parsed.Items) != 3 {
		t.Errorf("expected 3 items, got %d", len(parsed.Items))
	}

	expectedTotal := 100*2 + 150*1 + 200*3
	if parsed.Total != float64(expectedTotal) {
		t.Errorf("expected total %.2f, got %.2f", float64(expectedTotal), parsed.Total)
	}
}

func TestJSONExportEmptyCart(t *testing.T) {
	c := cart.Cart{Items: []cart.CartItem{}}

	v := &JSONExportVisitor{}
	result := ExportCart(c, v)

	var parsed jsonExportResult
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("Result is not valid JSON: %v\nGot: %s", err, result)
	}

	if len(parsed.Items) != 0 {
		t.Errorf("expected empty items array, got %d items", len(parsed.Items))
	}

	if parsed.Total != 0 {
		t.Errorf("expected total 0, got %.2f", parsed.Total)
	}
}

func TestTextReceiptSingleItem(t *testing.T) {
	c := cart.Cart{
		Items: []cart.CartItem{
			{ProductID: "p1", Name: "Royal Canin Dry", Price: 100, Quantity: 2},
		},
	}

	v := &TextReceiptVisitor{}
	result := ExportCart(c, v)

	if !strings.Contains(result, "=== Receipt ===") {
		t.Errorf("expected receipt header, got: %s", result)
	}

	if !strings.Contains(result, "Royal Canin Dry") {
		t.Errorf("expected item name in receipt, got: %s", result)
	}

	if !strings.Contains(result, "Total:") {
		t.Errorf("expected total line in receipt, got: %s", result)
	}
}

func TestTextReceiptMultipleItems(t *testing.T) {
	c := cart.Cart{
		Items: []cart.CartItem{
			{ProductID: "p1", Name: "Royal Canin Dry", Price: 100, Quantity: 2},
			{ProductID: "p2", Name: "Acana Adult", Price: 150, Quantity: 1},
		},
	}

	v := &TextReceiptVisitor{}
	result := ExportCart(c, v)

	if !strings.Contains(result, "Royal Canin Dry") {
		t.Errorf("expected first item name in receipt, got: %s", result)
	}

	if !strings.Contains(result, "Acana Adult") {
		t.Errorf("expected second item name in receipt, got: %s", result)
	}

	if !strings.Contains(result, "Total: 350.00") {
		t.Errorf("expected total 350.00 in receipt, got: %s", result)
	}
}

func TestTextReceiptEmptyCart(t *testing.T) {
	c := cart.Cart{Items: []cart.CartItem{}}

	v := &TextReceiptVisitor{}
	result := ExportCart(c, v)

	if !strings.Contains(result, "=== Receipt ===") {
		t.Errorf("expected receipt header, got: %s", result)
	}

	if !strings.Contains(result, "Total: 0.00") {
		t.Errorf("expected total 0.00 in receipt, got: %s", result)
	}
}

func TestExportCartConvenience(t *testing.T) {
	c := cart.Cart{
		Items: []cart.CartItem{
			{ProductID: "p1", Name: "Royal Canin Dry", Price: 100, Quantity: 2},
		},
	}

	jsonVisitor := &JSONExportVisitor{}
	jsonResult := ExportCart(c, jsonVisitor)
	if jsonResult == "" {
		t.Error("expected non-empty JSON result")
	}

	textVisitor := &TextReceiptVisitor{}
	textResult := ExportCart(c, textVisitor)
	if textResult == "" {
		t.Error("expected non-empty text result")
	}
}
