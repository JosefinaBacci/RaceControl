package sanctions

import (
	"time"

	"github.com/racecontrol/backend/internal/calendar"
	"github.com/racecontrol/backend/internal/db"
	"github.com/racecontrol/backend/internal/db/dbgen"
)

// Action is what a version did to its sanction. The set matches
// sanction_versions_action_valid in the schema.
type Action string

const (
	ActionIssued  Action = "issued"
	ActionAmended Action = "amended"
	ActionRevoked Action = "revoked"
)

// SubjectKind tells whether a sanction falls on a driver or on a team.
type SubjectKind string

const (
	SubjectDriver SubjectKind = "driver"
	SubjectTeam   SubjectKind = "team"
)

// PenaltyType is an entry of the catalogue. Unit and MaxValue are nil for a
// penalty that takes no value.
type PenaltyType struct {
	ID       int64
	Code     string
	Name     string
	Unit     *string
	MaxValue *int32
}

// Penalty is what a version imposes: a type and, when the type has a unit, its
// value.
type Penalty struct {
	Code  string
	Name  string
	Unit  *string
	Value *int32
}

// Subject is who the sanction falls on. The driver fields are nil for a team.
type Subject struct {
	Kind            SubjectKind
	DriverID        *int64
	DriverFirstName *string
	DriverLastName  *string
}

// Team is the accountable team: the sanctioned team, or the driver's team on the
// day of the sanction.
type Team struct {
	ID   int64
	Code string
	Name string
}

// EventRef is the calendar event a sanction was decided at.
type EventRef struct {
	ID       int64
	Name     string
	StartsAt time.Time
}

// Version is one step of a sanction's history. Note explains an amendment or a
// revocation and is nil for the issue.
type Version struct {
	ID          int64
	Action      Action
	Penalty     Penalty
	Reason      string
	Event       *EventRef
	Note        *string
	ChangedAt   time.Time
	ActorUserID int64
}

// Sanction is a sanction as it stands: its identity and its current version.
type Sanction struct {
	ID           int64
	Subject      Subject
	Team         Team
	CategoryCode string
	CategoryName string
	IssuedAt     time.Time
	Current      Version
}

// IsRevoked reports whether the version that stands is the revocation.
func (s Sanction) IsRevoked() bool {
	return s.Current.Action == ActionRevoked
}

// Detail is a sanction with its whole history, oldest version first.
type Detail struct {
	Sanction
	Versions []Version
}

// sanctionFromRow reads the sanction as of the version on the row.
func sanctionFromRow(row dbgen.SanctionVersionDetail) Sanction {
	return Sanction{
		ID:           row.SanctionID,
		Subject:      subjectFromRow(row),
		Team:         Team{ID: row.AccountableTeamID, Code: row.AccountableTeamCode, Name: row.AccountableTeamName},
		CategoryCode: row.CategoryCode,
		CategoryName: row.CategoryName,
		IssuedAt:     row.IssuedAt.Time,
		Current:      versionFromRow(row),
	}
}

func subjectFromRow(row dbgen.SanctionVersionDetail) Subject {
	if !row.DriverID.Valid {
		return Subject{Kind: SubjectTeam}
	}
	return Subject{
		Kind:            SubjectDriver,
		DriverID:        db.Int64Pointer(row.DriverID),
		DriverFirstName: db.TextPointer(row.DriverFirstName),
		DriverLastName:  db.TextPointer(row.DriverLastName),
	}
}

func versionFromRow(row dbgen.SanctionVersionDetail) Version {
	return Version{
		ID:     row.VersionID,
		Action: Action(row.Action),
		Penalty: Penalty{
			Code:  row.PenaltyCode,
			Name:  row.PenaltyName,
			Unit:  db.TextPointer(row.PenaltyUnit),
			Value: db.Int32Pointer(row.PenaltyValue),
		},
		Reason:      row.Reason,
		Event:       eventFromRow(row),
		Note:        db.TextPointer(row.Note),
		ChangedAt:   row.ChangedAt.Time,
		ActorUserID: row.ActorUserID,
	}
}

// eventFromRow labels the event with the calendar's own composition, so it reads
// the same here as in the calendar.
func eventFromRow(row dbgen.SanctionVersionDetail) *EventRef {
	if !row.EventID.Valid {
		return nil
	}
	label := calendar.Event{
		Kind:        calendar.Kind(row.EventKind.String),
		StartsAt:    row.EventStartsAt.Time,
		GPName:      db.TextPointer(row.EventGpName),
		RoundNumber: db.Int32Pointer(row.EventRoundNumber),
		ControlType: db.TextPointer(row.EventControlType),
	}
	return &EventRef{ID: row.EventID.Int64, Name: label.Name(), StartsAt: row.EventStartsAt.Time}
}
