package tyretests

import (
	"github.com/racecontrol/backend/internal/db"
	"github.com/racecontrol/backend/internal/db/dbgen"
	"github.com/racecontrol/backend/internal/results"
)

// Result is one official result of a tyre test for one compound, its type. A
// correction is a Result too, pointing at the one it supersedes.
type Result struct {
	results.Entry
	CircuitName string
	IsCurrent   bool
}

// NewResult is the first result of a compound in a test, as the FIA loads it.
type NewResult struct {
	results.Input
	CompoundCode string
}

// Correction replaces a result with a new one for the same test and compound,
// and says why the previous one was wrong.
type Correction struct {
	results.Input
	CorrectionNote string
}

func toResult(row dbgen.TyreTestResultEntry) Result {
	return Result{
		Entry: results.Entry{
			ID:             row.ID,
			EventID:        row.EventID,
			ConductedOn:    row.ConductedOn.Time,
			Type:           results.Type{Code: row.CompoundCode, Name: row.CompoundName},
			Passed:         row.Passed,
			Notes:          row.Notes,
			SupersedesID:   db.Int64Pointer(row.SupersedesID),
			CorrectionNote: db.TextPointer(row.CorrectionNote),
			SeasonYear:     row.SeasonYear,
			CategoryCode:   row.CategoryCode,
			RecordedBy:     row.RecordedByUsername,
			RecordedAt:     row.RecordedAt.Time,
		},
		CircuitName: row.CircuitName,
		IsCurrent:   row.IsCurrent,
	}
}

func toResults(rows []dbgen.TyreTestResultEntry) []Result {
	converted := make([]Result, 0, len(rows))
	for _, row := range rows {
		converted = append(converted, toResult(row))
	}
	return converted
}
