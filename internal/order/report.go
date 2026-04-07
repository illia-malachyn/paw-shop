package order

import (
	"fmt"
	"strings"
)

// ReportGenerator — інтерфейс для генерації звітів (Template Method pattern).
// Кожна реалізація визначає свої кроки Header, Body, Footer.
type ReportGenerator interface {
	Header() string
	Body(orders []*Order) string
	Footer() string
}

// GenerateReport — шаблонний метод генерації звіту.
// Викликає Header, Body, Footer у фіксованому порядку.
func GenerateReport(gen ReportGenerator, orders []*Order) string {
	return gen.Header() + "\n" + gen.Body(orders) + "\n" + gen.Footer()
}

// DailyReportGenerator — генератор короткого щоденного звіту.
type DailyReportGenerator struct{}

func (d *DailyReportGenerator) Header() string {
	return "=== Daily Report ==="
}

func (d *DailyReportGenerator) Body(orders []*Order) string {
	confirmed := 0
	rejected := 0
	for _, o := range orders {
		switch o.Status {
		case "confirmed":
			confirmed++
		case "rejected":
			rejected++
		}
	}
	return fmt.Sprintf("Total: %d | Confirmed: %d | Rejected: %d", len(orders), confirmed, rejected)
}

func (d *DailyReportGenerator) Footer() string {
	return "=== End of Daily Report ==="
}

// WeeklyReportGenerator — генератор детального тижневого звіту.
type WeeklyReportGenerator struct{}

func (w *WeeklyReportGenerator) Header() string {
	return "===== Weekly Report ====="
}

func (w *WeeklyReportGenerator) Body(orders []*Order) string {
	var b strings.Builder
	for i, o := range orders {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(fmt.Sprintf("Order %s: status=%s, items=%v", o.ID, o.Status, o.Items))
	}
	return b.String()
}

func (w *WeeklyReportGenerator) Footer() string {
	return "===== End of Weekly Report ====="
}
