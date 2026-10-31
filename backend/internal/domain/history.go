package domain

import (
	"strings"
	"time"
)

// HistoryZone is the zone a calendar day is counted in. Every history (technical
// controls, tyre tests, sanctions) filters and records days in it, so a day typed
// by the FIA and a day typed in a filter always mean the same 24 hours.
var HistoryZone = time.UTC

const dayLayout = time.DateOnly

// DateRange is a span of whole days. A nil bound is open. Until is exclusive: it
// is the start of the day after the last one asked for, so the last day counts in
// full whatever the column's time of day.
type DateRange struct {
	From  *time.Time
	Until *time.Time
}

// HistoryFilter narrows a history list by date range and by the kind of entry.
type HistoryFilter struct {
	Range DateRange
	Type  string
}

// ParseDay reads a YYYY-MM-DD day. A blank or impossible day (2026-02-30) is a
// validation error on field.
func ParseDay(field, value string) (time.Time, error) {
	day, err := time.ParseInLocation(dayLayout, strings.TrimSpace(value), HistoryZone)
	if err != nil {
		return time.Time{}, NewValidationError(field, "la fecha debe tener el formato AAAA-MM-DD")
	}
	return day, nil
}

// ParseDateRange reads the inclusive days from and to, either of them blank for
// an open bound.
func ParseDateRange(from, to string) (DateRange, error) {
	var dateRange DateRange
	if strings.TrimSpace(from) != "" {
		day, err := ParseDay("from", from)
		if err != nil {
			return DateRange{}, err
		}
		dateRange.From = &day
	}
	if strings.TrimSpace(to) != "" {
		day, err := ParseDay("to", to)
		if err != nil {
			return DateRange{}, err
		}
		until := day.AddDate(0, 0, 1)
		dateRange.Until = &until
	}
	if dateRange.From != nil && dateRange.Until != nil && !dateRange.From.Before(*dateRange.Until) {
		return DateRange{}, NewValidationError("to", "la fecha final no puede ser anterior a la inicial")
	}
	return dateRange, nil
}

// DayOf is the history day an instant falls on.
func DayOf(instant time.Time) time.Time {
	year, month, day := instant.In(HistoryZone).Date()
	return time.Date(year, month, day, 0, 0, 0, 0, HistoryZone)
}
