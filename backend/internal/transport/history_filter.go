package transport

import (
	"net/http"
	"strings"
	"time"

	"github.com/racecontrol/backend/internal/domain"
)

// dateRangeOf reads the inclusive days of the from and to query parameters.
func dateRangeOf(r *http.Request) (domain.DateRange, error) {
	query := r.URL.Query()
	return domain.ParseDateRange(query.Get("from"), query.Get("to"))
}

// historyFilterOf reads the date range and the type every history list filters by.
func historyFilterOf(r *http.Request) (domain.HistoryFilter, error) {
	dateRange, err := dateRangeOf(r)
	if err != nil {
		return domain.HistoryFilter{}, err
	}
	return domain.HistoryFilter{Range: dateRange, Type: strings.TrimSpace(r.URL.Query().Get("type"))}, nil
}

// dayResponse renders a history day as YYYY-MM-DD: a day has no time of day, so
// sending it as an instant would let a client shift it into the day before.
func dayResponse(day time.Time) string {
	return day.In(domain.HistoryZone).Format(time.DateOnly)
}
