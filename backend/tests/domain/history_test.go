package domaintest

import (
	"errors"
	"testing"
	"time"

	"github.com/racecontrol/backend/internal/domain"
)

func day(year int, month time.Month, dayOfMonth int) time.Time {
	return time.Date(year, month, dayOfMonth, 0, 0, 0, 0, time.UTC)
}

func TestParseDateRangeBounds(t *testing.T) {
	dateRange, err := domain.ParseDateRange("2026-03-01", "2026-03-31")
	if err != nil {
		t.Fatalf("ParseDateRange: %v", err)
	}
	if !dateRange.From.Equal(day(2026, 3, 1)) {
		t.Errorf("From = %v, want 2026-03-01", dateRange.From)
	}
	if !dateRange.Until.Equal(day(2026, 4, 1)) {
		t.Errorf("Until = %v, want the exclusive 2026-04-01", dateRange.Until)
	}
}

func TestParseDateRangeOpenBounds(t *testing.T) {
	dateRange, err := domain.ParseDateRange(" ", "")
	if err != nil {
		t.Fatalf("ParseDateRange: %v", err)
	}
	if dateRange.From != nil || dateRange.Until != nil {
		t.Errorf("blank bounds = %+v, want both open", dateRange)
	}
}

func TestParseDateRangeSingleDay(t *testing.T) {
	if _, err := domain.ParseDateRange("2026-03-01", "2026-03-01"); err != nil {
		t.Errorf("a one-day range was rejected: %v", err)
	}
}

func TestParseDateRangeRejects(t *testing.T) {
	cases := []struct {
		name      string
		from, to  string
		wantField string
	}{
		{name: "malformed from", from: "01/03/2026", wantField: "from"},
		{name: "impossible day", from: "2026-02-30", wantField: "from"},
		{name: "instant instead of a day", to: "2026-03-01T10:00:00Z", wantField: "to"},
		{name: "reversed range", from: "2026-03-02", to: "2026-03-01", wantField: "to"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := domain.ParseDateRange(tc.from, tc.to)
			var validationErr *domain.ValidationError
			if !errors.As(err, &validationErr) || validationErr.Field != tc.wantField {
				t.Fatalf("ParseDateRange(%q, %q) = %v, want a validation error on %q", tc.from, tc.to, err, tc.wantField)
			}
		})
	}
}

func TestDayOfCountsInTheHistoryZone(t *testing.T) {
	lateEvening := time.Date(2026, 3, 1, 22, 30, 0, 0, time.FixedZone("UTC-3", -3*60*60))
	if got := domain.DayOf(lateEvening); !got.Equal(day(2026, 3, 2)) {
		t.Errorf("DayOf(%v) = %v, want the UTC day 2026-03-02", lateEvening, got)
	}
}
