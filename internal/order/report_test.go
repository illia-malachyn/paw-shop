package order

import (
	"strings"
	"testing"
)

func TestGenerateReport(t *testing.T) {
	orders := []*Order{
		{ID: "o1", Status: "confirmed", Items: []string{"kibble", "treats"}},
		{ID: "o2", Status: "rejected", Items: []string{"wet-food"}},
	}

	t.Run("daily report contains header, body, footer", func(t *testing.T) {
		gen := &DailyReportGenerator{}
		result := GenerateReport(gen, orders)

		if !strings.Contains(result, "Daily Report") {
			t.Errorf("daily report missing 'Daily Report' header, got:\n%s", result)
		}
		if !strings.Contains(result, "Total:") {
			t.Errorf("daily report body missing 'Total:', got:\n%s", result)
		}
		if !strings.Contains(result, "End of Daily Report") {
			t.Errorf("daily report missing footer, got:\n%s", result)
		}
	})

	t.Run("weekly report contains header, per-order body, footer", func(t *testing.T) {
		gen := &WeeklyReportGenerator{}
		result := GenerateReport(gen, orders)

		if !strings.Contains(result, "Weekly Report") {
			t.Errorf("weekly report missing 'Weekly Report' header, got:\n%s", result)
		}
		if !strings.Contains(result, "o1") || !strings.Contains(result, "status=") {
			t.Errorf("weekly report body missing per-order details, got:\n%s", result)
		}
		if !strings.Contains(result, "End of Weekly Report") {
			t.Errorf("weekly report missing footer, got:\n%s", result)
		}
	})
}

func TestDailyReportGenerator(t *testing.T) {
	t.Run("body contains count summary", func(t *testing.T) {
		orders := []*Order{
			{ID: "o1", Status: "confirmed", Items: []string{"kibble"}},
			{ID: "o2", Status: "rejected", Items: []string{"wet-food"}},
		}
		gen := &DailyReportGenerator{}
		body := gen.Body(orders)

		if !strings.Contains(body, "Total:") {
			t.Errorf("daily body missing 'Total:', got: %q", body)
		}
		if !strings.Contains(body, "Confirmed:") {
			t.Errorf("daily body missing 'Confirmed:', got: %q", body)
		}
		if !strings.Contains(body, "Rejected:") {
			t.Errorf("daily body missing 'Rejected:', got: %q", body)
		}
	})

	t.Run("empty orders produces header and footer", func(t *testing.T) {
		gen := &DailyReportGenerator{}
		result := GenerateReport(gen, []*Order{})

		if !strings.Contains(result, "Daily Report") {
			t.Errorf("empty daily report missing header, got:\n%s", result)
		}
		if !strings.Contains(result, "End of Daily Report") {
			t.Errorf("empty daily report missing footer, got:\n%s", result)
		}
	})
}

func TestWeeklyReportGenerator(t *testing.T) {
	t.Run("body contains per-order details", func(t *testing.T) {
		orders := []*Order{
			{ID: "o1", Status: "confirmed", Items: []string{"kibble"}},
			{ID: "o2", Status: "new", Items: []string{"wet-food"}},
		}
		gen := &WeeklyReportGenerator{}
		body := gen.Body(orders)

		if !strings.Contains(body, "o1") {
			t.Errorf("weekly body missing order o1, got: %q", body)
		}
		if !strings.Contains(body, "o2") {
			t.Errorf("weekly body missing order o2, got: %q", body)
		}
		if !strings.Contains(body, "status=") {
			t.Errorf("weekly body missing 'status=', got: %q", body)
		}
	})

	t.Run("empty orders produces header and footer", func(t *testing.T) {
		gen := &WeeklyReportGenerator{}
		result := GenerateReport(gen, []*Order{})

		if !strings.Contains(result, "Weekly Report") {
			t.Errorf("empty weekly report missing header, got:\n%s", result)
		}
		if !strings.Contains(result, "End of Weekly Report") {
			t.Errorf("empty weekly report missing footer, got:\n%s", result)
		}
	})

	t.Run("body differs from daily body for same input", func(t *testing.T) {
		orders := []*Order{
			{ID: "o1", Status: "confirmed", Items: []string{"kibble"}},
		}
		daily := &DailyReportGenerator{}
		weekly := &WeeklyReportGenerator{}

		dailyBody := daily.Body(orders)
		weeklyBody := weekly.Body(orders)

		if dailyBody == weeklyBody {
			t.Error("DailyReportGenerator and WeeklyReportGenerator Body() produce identical output, expected different")
		}
	})
}
