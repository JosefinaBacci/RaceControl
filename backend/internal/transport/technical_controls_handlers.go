package transport

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/racecontrol/backend/internal/domain"
	"github.com/racecontrol/backend/internal/results"
	"github.com/racecontrol/backend/internal/technicalcontrols"
)

type TechnicalControlBoard interface {
	History(ctx context.Context, filter technicalcontrols.Filter) ([]technicalcontrols.Result, error)
	Types(ctx context.Context) ([]results.Type, error)
	Sheet(ctx context.Context, eventID int64) (technicalcontrols.Sheet, error)
	Record(ctx context.Context, input technicalcontrols.NewResult) (technicalcontrols.Result, error)
	Correct(ctx context.Context, correction technicalcontrols.Correction) (technicalcontrols.Result, error)
}

type technicalControlHandlers struct {
	controls TechnicalControlBoard
}

type gradedTeamResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type controlResultResponse struct {
	resultEntryResponse
	Team gradedTeamResponse `json:"team"`
}

type controlRevisionResponse struct {
	resultRevisionResponse
	Team gradedTeamResponse `json:"team"`
}

type controlHistoryResponse struct {
	Results []controlResultResponse `json:"results"`
}

type controlSheetTeamResponse struct {
	ID        int64                     `json:"id"`
	Name      string                    `json:"name"`
	IsActive  bool                      `json:"isActive"`
	Revisions []controlRevisionResponse `json:"revisions"`
}

type controlResponse struct {
	EventID      int64     `json:"eventId"`
	ControlType  string    `json:"controlType"`
	StartsAt     time.Time `json:"startsAt"`
	EndsAt       time.Time `json:"endsAt"`
	IsActive     bool      `json:"isActive"`
	SeasonYear   int32     `json:"seasonYear"`
	CategoryCode string    `json:"categoryCode"`
	CategoryName string    `json:"categoryName"`
}

type controlSheetResponse struct {
	Control controlResponse            `json:"control"`
	Teams   []controlSheetTeamResponse `json:"teams"`
}

type recordControlResultRequest struct {
	resultInputRequest
	TeamID int64 `json:"teamId"`
}

func (h *technicalControlHandlers) history(w http.ResponseWriter, r *http.Request) {
	filter, err := controlFilterOf(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	history, err := h.controls.History(r.Context(), filter)
	if err != nil {
		writeError(w, r, err)
		return
	}
	response := controlHistoryResponse{Results: make([]controlResultResponse, 0, len(history))}
	for _, result := range history {
		response.Results = append(response.Results, controlResultResponse{
			resultEntryResponse: toResultEntryResponse(result.Entry),
			Team:                toGradedTeamResponse(result.Team),
		})
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *technicalControlHandlers) types(w http.ResponseWriter, r *http.Request) {
	types, err := h.controls.Types(r.Context())
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toHistoryTypeListResponse(types))
}

func (h *technicalControlHandlers) sheet(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	sheet, err := h.controls.Sheet(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toControlSheetResponse(sheet))
}

func (h *technicalControlHandlers) record(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	actor, err := identityFrom(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	var request recordControlResultRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, r, err)
		return
	}
	result, err := h.controls.Record(r.Context(), technicalcontrols.NewResult{
		EventID:    id,
		TeamID:     request.TeamID,
		Input:      request.toInput(),
		RecordedBy: actor.UserID,
	})
	h.respondWithRevision(w, r, result, err)
}

func (h *technicalControlHandlers) correct(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	actor, err := identityFrom(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	var request resultCorrectionRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, r, err)
		return
	}
	result, err := h.controls.Correct(r.Context(), technicalcontrols.Correction{
		ResultID:   id,
		Input:      request.toInput(),
		Note:       request.CorrectionNote,
		RecordedBy: actor.UserID,
	})
	h.respondWithRevision(w, r, result, err)
}

func (h *technicalControlHandlers) respondWithRevision(w http.ResponseWriter, r *http.Request, result technicalcontrols.Result, err error) {
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toControlRevisionResponse(result))
}

// controlFilterOf adds the championship and the team to the shared history
// filter. A team that is not a positive id is a malformed filter, not an
// unknown team, so it is refused instead of matching nothing.
func controlFilterOf(r *http.Request) (technicalcontrols.Filter, error) {
	history, err := historyFilterOf(r)
	if err != nil {
		return technicalcontrols.Filter{}, err
	}
	filter := technicalcontrols.Filter{History: history, Category: strings.TrimSpace(r.URL.Query().Get("category"))}
	team := strings.TrimSpace(r.URL.Query().Get("team"))
	if team == "" {
		return filter, nil
	}
	filter.TeamID, err = strconv.ParseInt(team, 10, 64)
	if err != nil || filter.TeamID <= 0 {
		return technicalcontrols.Filter{}, domain.NewValidationError("team", "la escudería debe ser un id numérico")
	}
	return filter, nil
}

func toGradedTeamResponse(team technicalcontrols.Team) gradedTeamResponse {
	return gradedTeamResponse{ID: team.ID, Name: team.Name}
}

func toControlRevisionResponse(result technicalcontrols.Result) controlRevisionResponse {
	return controlRevisionResponse{
		resultRevisionResponse: toResultRevisionResponse(result.Entry),
		Team:                   toGradedTeamResponse(result.Team),
	}
}

func toControlSheetResponse(sheet technicalcontrols.Sheet) controlSheetResponse {
	control := sheet.Control
	response := controlSheetResponse{
		Control: controlResponse{
			EventID:      control.EventID,
			ControlType:  control.ControlType,
			StartsAt:     control.StartsAt,
			EndsAt:       control.EndsAt,
			IsActive:     control.IsActive,
			SeasonYear:   control.SeasonYear,
			CategoryCode: control.CategoryCode,
			CategoryName: control.CategoryName,
		},
		Teams: make([]controlSheetTeamResponse, 0, len(sheet.Teams)),
	}
	for _, team := range sheet.Teams {
		revisions := make([]controlRevisionResponse, 0, len(team.Revisions))
		for _, revision := range team.Revisions {
			revisions = append(revisions, toControlRevisionResponse(revision))
		}
		response.Teams = append(response.Teams, controlSheetTeamResponse{
			ID: team.ID, Name: team.Name, IsActive: team.IsActive, Revisions: revisions,
		})
	}
	return response
}
