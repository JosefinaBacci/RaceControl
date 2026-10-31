package transport

import (
	"time"

	"github.com/racecontrol/backend/internal/results"
)

// The JSON every official-results history shares (technical controls, tyre
// tests). Each feature embeds these and adds the subject it grades.

type historyTypeResponse struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type historyTypeListResponse struct {
	Types []historyTypeResponse `json:"types"`
}

type recorderResponse struct {
	Username string `json:"username"`
}

// resultEntryResponse is the public view: it never names the administrator who
// recorded it, since a public list of admin usernames helps credential attacks.
type resultEntryResponse struct {
	ID             int64               `json:"id"`
	EventID        int64               `json:"eventId"`
	ConductedOn    string              `json:"conductedOn"`
	Type           historyTypeResponse `json:"type"`
	Passed         bool                `json:"passed"`
	Notes          string              `json:"notes"`
	IsCorrection   bool                `json:"isCorrection"`
	CorrectionNote *string             `json:"correctionNote"`
	SeasonYear     int32               `json:"seasonYear"`
	CategoryCode   string              `json:"categoryCode"`
	RecordedAt     time.Time           `json:"recordedAt"`
}

// resultRevisionResponse is the FIA view of the same entry, with its author and
// the entry it replaced.
type resultRevisionResponse struct {
	resultEntryResponse
	SupersedesID *int64           `json:"supersedesId"`
	RecordedBy   recorderResponse `json:"recordedBy"`
}

type resultInputRequest struct {
	ConductedOn string `json:"conductedOn"`
	Passed      *bool  `json:"passed"`
	Notes       string `json:"notes"`
}

type resultCorrectionRequest struct {
	resultInputRequest
	CorrectionNote string `json:"correctionNote"`
}

func (r resultInputRequest) toInput() results.Input {
	return results.Input{ConductedOn: r.ConductedOn, Passed: r.Passed, Notes: r.Notes}
}

func toResultEntryResponse(entry results.Entry) resultEntryResponse {
	return resultEntryResponse{
		ID:             entry.ID,
		EventID:        entry.EventID,
		ConductedOn:    dayResponse(entry.ConductedOn),
		Type:           toHistoryTypeResponse(entry.Type),
		Passed:         entry.Passed,
		Notes:          entry.Notes,
		IsCorrection:   entry.IsCorrection(),
		CorrectionNote: entry.CorrectionNote,
		SeasonYear:     entry.SeasonYear,
		CategoryCode:   entry.CategoryCode,
		RecordedAt:     entry.RecordedAt.UTC(),
	}
}

func toResultRevisionResponse(entry results.Entry) resultRevisionResponse {
	return resultRevisionResponse{
		resultEntryResponse: toResultEntryResponse(entry),
		SupersedesID:        entry.SupersedesID,
		RecordedBy:          recorderResponse{Username: entry.RecordedBy},
	}
}

func toHistoryTypeResponse(historyType results.Type) historyTypeResponse {
	return historyTypeResponse{Code: historyType.Code, Name: historyType.Name}
}

func toHistoryTypeListResponse(types []results.Type) historyTypeListResponse {
	response := historyTypeListResponse{Types: make([]historyTypeResponse, 0, len(types))}
	for _, historyType := range types {
		response.Types = append(response.Types, toHistoryTypeResponse(historyType))
	}
	return response
}
