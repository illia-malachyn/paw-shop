package search

import (
	"strings"
	"testing"
)

// Тестові дані товарів на основі реального каталогу.
var (
	royal = ProductData{ID: "rc-dry-01", Name: "Royal Canin Maxi Adult", Brand: "Royal Canin", Category: "dry", Price: 1450}
	treat = ProductData{ID: "rc-treat-01", Name: "Royal Canin Dental Sticks", Brand: "Royal Canin", Category: "treat", Price: 280}
	acana = ProductData{ID: "ac-dry-01", Name: "Acana Wild Prairie", Brand: "Acana", Category: "dry", Price: 1850}
)

func TestBrandExpression(t *testing.T) {
	cases := []struct {
		name    string
		expr    BrandExpression
		product ProductData
		want    bool
	}{
		{"partial match royal", BrandExpression{Brand: "Royal"}, royal, true},
		{"full brand match", BrandExpression{Brand: "Royal Canin"}, royal, true},
		{"case insensitive lower", BrandExpression{Brand: "royal canin"}, royal, true},
		{"case insensitive upper", BrandExpression{Brand: "ACANA"}, acana, true},
		{"partial brand acana", BrandExpression{Brand: "Acana"}, acana, true},
		{"no match different brand", BrandExpression{Brand: "Unknown"}, royal, false},
		{"no match royal vs acana", BrandExpression{Brand: "Royal"}, acana, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.expr.Interpret(tc.product)
			if got != tc.want {
				t.Errorf("BrandExpression{%q}.Interpret(%v) = %v, want %v", tc.expr.Brand, tc.product.Brand, got, tc.want)
			}
		})
	}
}

func TestPriceLessThanExpression(t *testing.T) {
	cases := []struct {
		name    string
		expr    PriceLessThanExpression
		product ProductData
		want    bool
	}{
		{"price below threshold", PriceLessThanExpression{Price: 500}, treat, true},
		{"price equal threshold", PriceLessThanExpression{Price: 280}, treat, false},
		{"price above threshold", PriceLessThanExpression{Price: 100}, treat, false},
		{"price well below", PriceLessThanExpression{Price: 2000}, royal, true},
		{"price exactly at 1450", PriceLessThanExpression{Price: 1450}, royal, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.expr.Interpret(tc.product)
			if got != tc.want {
				t.Errorf("PriceLessThanExpression{%.0f}.Interpret(Price=%.0f) = %v, want %v", tc.expr.Price, tc.product.Price, got, tc.want)
			}
		})
	}
}

func TestCategoryExpression(t *testing.T) {
	cases := []struct {
		name    string
		expr    CategoryExpression
		product ProductData
		want    bool
	}{
		{"exact match dry", CategoryExpression{Category: "dry"}, royal, true},
		{"case insensitive DRY", CategoryExpression{Category: "DRY"}, royal, true},
		{"case insensitive Dry", CategoryExpression{Category: "Dry"}, acana, true},
		{"exact match treat", CategoryExpression{Category: "treat"}, treat, true},
		{"no match wet vs dry", CategoryExpression{Category: "wet"}, royal, false},
		{"no match treat vs dry", CategoryExpression{Category: "treat"}, royal, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.expr.Interpret(tc.product)
			if got != tc.want {
				t.Errorf("CategoryExpression{%q}.Interpret(Category=%q) = %v, want %v", tc.expr.Category, tc.product.Category, got, tc.want)
			}
		})
	}
}

func TestAndExpression(t *testing.T) {
	cases := []struct {
		name    string
		left    Expression
		right   Expression
		product ProductData
		want    bool
	}{
		{
			"both true: Royal AND price<500",
			BrandExpression{Brand: "Royal"},
			PriceLessThanExpression{Price: 500},
			treat,
			true,
		},
		{
			"left false: Unknown brand",
			BrandExpression{Brand: "Unknown"},
			PriceLessThanExpression{Price: 500},
			treat,
			false,
		},
		{
			"right false: price too high",
			BrandExpression{Brand: "Royal"},
			PriceLessThanExpression{Price: 500},
			royal,
			false,
		},
		{
			"both false",
			BrandExpression{Brand: "Unknown"},
			PriceLessThanExpression{Price: 100},
			royal,
			false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expr := AndExpression{Left: tc.left, Right: tc.right}
			got := expr.Interpret(tc.product)
			if got != tc.want {
				t.Errorf("AndExpression.Interpret(%v) = %v, want %v", tc.product, got, tc.want)
			}
		})
	}
}

func TestParse(t *testing.T) {
	cases := []struct {
		name    string
		query   string
		product ProductData
		want    bool
	}{
		{
			"single brand matches royal",
			"brand:Royal",
			royal,
			true,
		},
		{
			"single brand no match acana",
			"brand:Royal",
			acana,
			false,
		},
		{
			"single price matches treat",
			"price:<500",
			treat,
			true,
		},
		{
			"single price no match royal",
			"price:<500",
			royal,
			false,
		},
		{
			"single category matches dry",
			"category:dry",
			royal,
			true,
		},
		{
			"single category no match treat",
			"category:dry",
			treat,
			false,
		},
		{
			"AND brand and price matches treat",
			"brand:Royal AND price:<500",
			treat,
			true,
		},
		{
			"AND brand and price no match royal (too expensive)",
			"brand:Royal AND price:<500",
			royal,
			false,
		},
		{
			"triple AND matches treat",
			"brand:Royal AND price:<500 AND category:treat",
			treat,
			true,
		},
		{
			"triple AND no match royal (wrong category)",
			"brand:Royal AND price:<500 AND category:treat",
			royal,
			false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expr, err := Parse(tc.query)
			if err != nil {
				t.Fatalf("Parse(%q) returned unexpected error: %v", tc.query, err)
			}
			if expr == nil {
				t.Fatalf("Parse(%q) returned nil expression", tc.query)
			}
			got := expr.Interpret(tc.product)
			if got != tc.want {
				t.Errorf("Parse(%q).Interpret(%v) = %v, want %v", tc.query, tc.product, got, tc.want)
			}
		})
	}
}

func TestParseErrors(t *testing.T) {
	cases := []struct {
		name  string
		query string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
		{"unknown prefix", "invalid"},
		{"unknown key", "unknown:value"},
		{"invalid price number", "price:<abc"},
		{"missing brand value", "brand:"},
		{"missing category value", "category:"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expr, err := Parse(tc.query)
			if err == nil {
				t.Errorf("Parse(%q) expected error, got expression: %v", tc.query, expr)
			}
		})
	}
}

func TestParseDepthGuard(t *testing.T) {
	// Construct a query with 33 brand tokens joined by AND (exceeds maxDepth=32).
	tokens := make([]string, 33)
	for i := range tokens {
		tokens[i] = "brand:X"
	}
	query := strings.Join(tokens, " AND ")

	_, err := Parse(query)
	if err == nil {
		t.Fatal("Parse with 33 AND clauses expected error, got nil")
	}
	if !strings.Contains(err.Error(), "too complex") {
		t.Errorf("expected error to contain 'too complex', got: %v", err)
	}
}
