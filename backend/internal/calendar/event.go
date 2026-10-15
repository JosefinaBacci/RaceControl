package calendar

import (
	"fmt"
	"time"

	"github.com/racecontrol/backend/internal/db"
	"github.com/racecontrol/backend/internal/db/dbgen"
)

// Kind is the type of an event. The set matches events_kind_valid in the schema.
type Kind string

const (
	KindRace             Kind = "race"
	KindTyreTest         Kind = "tyre_test"
	KindTechnicalControl Kind = "technical_control"
)

// IsValid reports whether the kind is one the schema accepts.
func (k Kind) IsValid() bool {
	switch k {
	case KindRace, KindTyreTest, KindTechnicalControl:
		return true
	}
	return false
}

// tyreTestText is the fixed text a tyre test gets: nothing but its date
// distinguishes one from another, so the label carries both.
const tyreTestText = "Prueba de neumáticos"

const labelDateLayout = "02/01/2006"

// Event is one row of the calendar: the event plus the championship of whichever
// child it has. The attributes only one kind carries are nil for the others.
type Event struct {
	ID            int64
	Kind          Kind
	StartsAt      time.Time
	EndsAt        time.Time
	IsActive      bool
	DeactivatedAt *time.Time

	RoundID     *int64
	RoundNumber *int32
	GPName      *string
	ControlType *string

	SeasonID     int64
	SeasonYear   int32
	CategoryID   int64
	CategoryCode string
	CategoryName string

	CircuitID   *int64
	CircuitName *string
	CountryCode *string
}

// Name composes the label from what identifies the event by kind: a race names the
// Grand Prix of its round and the round number, a tyre test is its fixed text and
// date, and a control is its control type. Nothing is stored, so the label cannot
// drift from the data.
func (e Event) Name() string {
	switch e.Kind {
	case KindRace:
		return raceName(e.GPName, e.RoundNumber)
	case KindTyreTest:
		return fmt.Sprintf("%s · %s", tyreTestText, e.StartsAt.UTC().Format(labelDateLayout))
	case KindTechnicalControl:
		return textOf(e.ControlType)
	default:
		return ""
	}
}

func raceName(gpName *string, number *int32) string {
	if number == nil {
		return textOf(gpName)
	}
	return fmt.Sprintf("%s · Ronda %d", textOf(gpName), *number)
}

// Circuit is a catalogue entry the race form picks from.
type Circuit struct {
	ID          int64
	Name        string
	Locality    string
	CountryCode string
}

// Season is the championship year an event belongs to.
type Season struct {
	ID   int64
	Year int32
}

func eventFromRow(row dbgen.CalendarEvent) Event {
	return Event{
		ID:            row.ID,
		Kind:          Kind(row.Kind),
		StartsAt:      row.StartsAt.Time,
		EndsAt:        row.EndsAt.Time,
		IsActive:      row.IsActive,
		DeactivatedAt: db.TimePointer(row.DeactivatedAt),
		RoundID:       db.Int64Pointer(row.RoundID),
		RoundNumber:   db.Int32Pointer(row.RoundNumber),
		GPName:        db.TextPointer(row.GpName),
		ControlType:   db.TextPointer(row.ControlType),
		SeasonID:      row.SeasonID,
		SeasonYear:    row.SeasonYear,
		CategoryID:    row.CategoryID,
		CategoryCode:  row.CategoryCode,
		CategoryName:  row.CategoryName,
		CircuitID:     db.Int64Pointer(row.CircuitID),
		CircuitName:   db.TextPointer(row.CircuitName),
		CountryCode:   db.TextPointer(row.CountryCode),
	}
}

func textOf(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
