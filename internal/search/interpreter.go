// Package search реалізує патерн Interpreter для розбору та виконання текстових пошукових запитів.
// Підтримуваний синтаксис: brand:Royal, price:<500, category:dry, та їх комбінації через AND.
package search

import (
	"fmt"
	"strconv"
	"strings"
)

// maxDepth — максимальна глибина вкладеності AND-виразів.
// Захищає від DoS через надмірно складні запити (CVE-2024-34155 precedent).
const maxDepth = 32

// ProductData — дані товару для інтерпретації виразу.
// Відокремлено від models.Product, щоб пошук не залежав від HTTP-шару.
type ProductData struct {
	ID       string
	Name     string
	Brand    string
	Category string
	Price    float64
}

// Expression — інтерфейс патерну Interpreter.
// Кожен вираз може бути обчислений відносно конкретного товару.
type Expression interface {
	Interpret(product ProductData) bool
}

// BrandExpression — термінальний вираз для фільтрації за брендом.
// Перевірка нечутлива до регістру та підтримує часткове співпадіння.
type BrandExpression struct {
	Brand string
}

// Interpret — повертає true, якщо бренд товару містить рядок e.Brand (нечутливо до регістру).
func (e BrandExpression) Interpret(product ProductData) bool {
	return strings.Contains(strings.ToLower(product.Brand), strings.ToLower(e.Brand))
}

// PriceLessThanExpression — термінальний вираз для фільтрації за ціною.
// Перевіряє, що ціна товару є строго меншою за порогове значення.
type PriceLessThanExpression struct {
	Price float64
}

// Interpret — повертає true, якщо ціна товару менша за e.Price.
func (e PriceLessThanExpression) Interpret(product ProductData) bool {
	return product.Price < e.Price
}

// CategoryExpression — термінальний вираз для фільтрації за категорією.
// Перевірка нечутлива до регістру.
type CategoryExpression struct {
	Category string
}

// Interpret — повертає true, якщо категорія товару відповідає e.Category (нечутливо до регістру).
func (e CategoryExpression) Interpret(product ProductData) bool {
	return strings.EqualFold(product.Category, e.Category)
}

// AndExpression — нетермінальний вираз, який об'єднує два вирази логічним AND.
type AndExpression struct {
	Left  Expression
	Right Expression
}

// Interpret — повертає true, якщо обидва вирази є істинними для даного товару.
func (e AndExpression) Interpret(product ProductData) bool {
	return e.Left.Interpret(product) && e.Right.Interpret(product)
}

// Parse розбирає рядок запиту в дерево виразів Expression.
// Підтримуваний синтаксис: brand:X, price:<N, category:X, та їх комбінації через AND.
// Повертає помилку для порожнього рядка, некоректного формату або надмірно складного запиту.
func Parse(query string) (Expression, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	tokens := strings.Split(query, " AND ")

	if len(tokens) > maxDepth {
		return nil, fmt.Errorf("query too complex: exceeds maximum depth of %d", maxDepth)
	}

	exprs := make([]Expression, 0, len(tokens))
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		expr, err := parseToken(token)
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, expr)
	}

	if len(exprs) == 0 {
		return nil, fmt.Errorf("no valid expressions found in query")
	}

	// Combine left-to-right with AndExpression (left-associative).
	result := exprs[0]
	for i := 1; i < len(exprs); i++ {
		result = AndExpression{Left: result, Right: exprs[i]}
	}

	return result, nil
}

// parseToken розбирає окремий токен запиту в конкретний Expression.
func parseToken(token string) (Expression, error) {
	if strings.HasPrefix(token, "brand:") {
		brand := strings.TrimPrefix(token, "brand:")
		if brand == "" {
			return nil, fmt.Errorf("brand value cannot be empty in token: %q", token)
		}
		return BrandExpression{Brand: brand}, nil
	}

	if strings.HasPrefix(token, "price:<") {
		priceStr := strings.TrimPrefix(token, "price:<")
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid price value %q: %w", priceStr, err)
		}
		return PriceLessThanExpression{Price: price}, nil
	}

	if strings.HasPrefix(token, "category:") {
		category := strings.TrimPrefix(token, "category:")
		if category == "" {
			return nil, fmt.Errorf("category value cannot be empty in token: %q", token)
		}
		return CategoryExpression{Category: category}, nil
	}

	return nil, fmt.Errorf("unknown expression format: %q", token)
}
