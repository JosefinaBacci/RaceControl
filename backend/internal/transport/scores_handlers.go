package transport

import (
	"context"
	"net/http"
	"strconv"

	"github.com/racecontrol/backend/internal/domain"
	"github.com/racecontrol/backend/internal/scores"
)

type ScoreManager interface {
	Grid(ctx context.Context, eventID int64) (scores.Grid, error)
	Replace(ctx context.Context, eventID, actorUserID int64, entries []scores.Entry) (scores.Grid, error)
	DriverStandings(ctx context.Context, category string, seasonYear int32) ([]scores.Standing, error)
}

type scoreHandlers struct {
	scores ScoreManager
}

type scoreGridEntryResponse struct {
	DriverID    int64    `json:"driverId"`
	FirstName   string   `json:"firstName"`
	LastName    string   `json:"lastName"`
	CountryCode string   `json:"countryCode"`
	TeamID      int64    `json:"teamId"`
	TeamName    string   `json:"teamName"`
	Number      *int32   `json:"number"`
	Points      *float64 `json:"points"`
}

type scoreGridResponse struct {
	EventID      int64                    `json:"eventId"`
	CategoryCode string                   `json:"categoryCode"`
	CategoryName string                   `json:"categoryName"`
	Entries      []scoreGridEntryResponse `json:"entries"`
}

type replaceScoresRequest struct {
	Entries []scoreEntryRequest `json:"entries"`
}

type scoreEntryRequest struct {
	DriverID int64    `json:"driverId"`
	Points   *float64 `json:"points"`
}

type standingResponse struct {
	DriverID    int64   `json:"driverId"`
	FirstName   string  `json:"firstName"`
	LastName    string  `json:"lastName"`
	CountryCode string  `json:"countryCode"`
	TeamID      int64   `json:"teamId"`
	TeamName    string  `json:"teamName"`
	Points      float64 `json:"points"`
}

type standingsResponse struct {
	Standings []standingResponse `json:"standings"`
}

func (h *scoreHandlers) grid(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	grid, err := h.scores.Grid(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toScoreGridResponse(grid))
}

func (h *scoreHandlers) replace(w http.ResponseWriter, r *http.Request) {
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
	var request replaceScoresRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, r, err)
		return
	}
	entries := make([]scores.Entry, 0, len(request.Entries))
	for _, entry := range request.Entries {
		entries = append(entries, scores.Entry{DriverID: entry.DriverID, Points: entry.Points})
	}
	grid, err := h.scores.Replace(r.Context(), id, actor.UserID, entries)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toScoreGridResponse(grid))
}

func (h *scoreHandlers) standings(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	category := query.Get("category")
	if category == "" {
		writeError(w, r, domain.NewValidationError("category", "la categoría es obligatoria"))
		return
	}
	seasonYear, err := strconv.Atoi(query.Get("season"))
	if err != nil {
		writeError(w, r, domain.NewValidationError("season", "la temporada es obligatoria"))
		return
	}
	standings, err := h.scores.DriverStandings(r.Context(), category, int32(seasonYear))
	if err != nil {
		writeError(w, r, err)
		return
	}
	response := standingsResponse{Standings: make([]standingResponse, 0, len(standings))}
	for _, standing := range standings {
		response.Standings = append(response.Standings, standingResponse(standing))
	}
	writeJSON(w, http.StatusOK, response)
}

func toScoreGridResponse(grid scores.Grid) scoreGridResponse {
	response := scoreGridResponse{
		EventID:      grid.EventID,
		CategoryCode: grid.CategoryCode,
		CategoryName: grid.CategoryName,
		Entries:      make([]scoreGridEntryResponse, 0, len(grid.Entries)),
	}
	for _, entry := range grid.Entries {
		response.Entries = append(response.Entries, scoreGridEntryResponse{
			DriverID:    entry.DriverID,
			FirstName:   entry.FirstName,
			LastName:    entry.LastName,
			CountryCode: entry.CountryCode,
			TeamID:      entry.TeamID,
			TeamName:    entry.TeamName,
			Number:      entry.Number,
			Points:      entry.Points,
		})
	}
	return response
}
