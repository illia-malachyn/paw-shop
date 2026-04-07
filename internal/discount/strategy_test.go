package discount

import (
	"math"
	"testing"
)

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.01
}

func TestPercentStrategy(t *testing.T) {
	s := &PercentStrategy{Percent: 10}
	if s.Name() != "percent" {
		t.Errorf("expected name 'percent', got '%s'", s.Name())
	}
	result := s.Apply(1000)
	if !almostEqual(result, 900) {
		t.Errorf("expected 900, got %.2f", result)
	}
}

func TestPercentStrategy_50(t *testing.T) {
	s := &PercentStrategy{Percent: 50}
	result := s.Apply(200)
	if !almostEqual(result, 100) {
		t.Errorf("expected 100, got %.2f", result)
	}
}

func TestFixedStrategy(t *testing.T) {
	s := &FixedStrategy{Amount: 150}
	if s.Name() != "fixed" {
		t.Errorf("expected name 'fixed', got '%s'", s.Name())
	}
	result := s.Apply(500)
	if !almostEqual(result, 350) {
		t.Errorf("expected 350, got %.2f", result)
	}
}

func TestFixedStrategy_ExceedsPrice(t *testing.T) {
	s := &FixedStrategy{Amount: 600}
	result := s.Apply(500)
	if !almostEqual(result, 0) {
		t.Errorf("expected 0 (floor), got %.2f", result)
	}
}

func TestBuyNGetOneStrategy(t *testing.T) {
	s := &BuyNGetOneStrategy{N: 2}
	if s.Name() != "buy_n_get_one" {
		t.Errorf("expected name 'buy_n_get_one', got '%s'", s.Name())
	}
	// Купуєш 2, отримуєш 3. Платиш за 2/3 ціни.
	result := s.Apply(300)
	if !almostEqual(result, 200) {
		t.Errorf("expected 200, got %.2f", result)
	}
}
