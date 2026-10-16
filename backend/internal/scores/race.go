package scores

import (
	"time"

	"github.com/racecontrol/backend/internal/calendar"
	"github.com/racecontrol/backend/internal/db"
	"github.com/racecontrol/backend/internal/db/dbgen"
)

// Race identifies the race a result or a score change belongs to.
type Race struct {
	ID           int64
	Name         string
	GPName       *string
	RoundNumber  *int32
	StartsAt     time.Time
	SeasonYear   int32
	CategoryCode string
	CategoryName string
	CircuitName  *string
	CountryCode  *string
}

// raceFromEvent labels the race with the calendar's own composition, so a race
// reads the same in the calendar, the results and the notices.
func raceFromEvent(event dbgen.CalendarEvent) Race {
	gpName := db.TextPointer(event.GpName)
	roundNumber := db.Int32Pointer(event.RoundNumber)
	return Race{
		ID:           event.ID,
		Name:         calendar.Event{Kind: calendar.KindRace, GPName: gpName, RoundNumber: roundNumber}.Name(),
		GPName:       gpName,
		RoundNumber:  roundNumber,
		StartsAt:     event.StartsAt.Time,
		SeasonYear:   event.SeasonYear,
		CategoryCode: event.CategoryCode,
		CategoryName: event.CategoryName,
		CircuitName:  db.TextPointer(event.CircuitName),
		CountryCode:  db.TextPointer(event.CountryCode),
	}
}
