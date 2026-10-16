package transport

import (
	"context"
	"net/http"
	"time"

	"github.com/racecontrol/backend/internal/scores"
)

type ScoreBoard interface {
	Results(ctx context.Context, filter scores.Filter) ([]scores.RaceResults, error)
	RaceResults(ctx context.Context, eventID int64) (scores.RaceResults, error)
	Standings(ctx context.Context, filter scores.Filter) (scores.Standings, error)
}

type scoreHandlers struct {
	scores ScoreBoard
}

type raceResponse struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	GPName       *string   `json:"gpName"`
	RoundNumber  *int32    `json:"roundNumber"`
	StartsAt     time.Time `json:"startsAt"`
	SeasonYear   int32     `json:"seasonYear"`
	CategoryCode string    `json:"categoryCode"`
	CategoryName string    `json:"categoryName"`
	CircuitName  *string   `json:"circuitName"`
	CountryCode  *string   `json:"countryCode"`
}

type driverResultResponse struct {
	Position    int32   `json:"position"`
	DriverID    int64   `json:"driverId"`
	FirstName   string  `json:"firstName"`
	LastName    string  `json:"lastName"`
	CountryCode string  `json:"countryCode"`
	Number      *int32  `json:"number"`
	TeamID      int64   `json:"teamId"`
	TeamCode    string  `json:"teamCode"`
	TeamName    string  `json:"teamName"`
	Points      float64 `json:"points"`
}

type raceResultsResponse struct {
	Race    raceResponse           `json:"race"`
	Results []driverResultResponse `json:"results"`
}

type raceResultsListResponse struct {
	Races []raceResultsResponse `json:"races"`
}

type driverStandingResponse struct {
	Position    int32   `json:"position"`
	DriverID    int64   `json:"driverId"`
	FirstName   string  `json:"firstName"`
	LastName    string  `json:"lastName"`
	CountryCode string  `json:"countryCode"`
	Number      *int32  `json:"number"`
	TeamID      int64   `json:"teamId"`
	TeamCode    string  `json:"teamCode"`
	TeamName    string  `json:"teamName"`
	Points      float64 `json:"points"`
	Races       int32   `json:"races"`
}

type teamStandingResponse struct {
	Position int32   `json:"position"`
	TeamID   int64   `json:"teamId"`
	TeamCode string  `json:"teamCode"`
	TeamName string  `json:"teamName"`
	Points   float64 `json:"points"`
}

type standingsResponse struct {
	SeasonYear   int32                    `json:"seasonYear"`
	CategoryCode string                   `json:"categoryCode"`
	Drivers      []driverStandingResponse `json:"drivers"`
	Teams        []teamStandingResponse   `json:"teams"`
}

func (h *scoreHandlers) results(w http.ResponseWriter, r *http.Request) {
	races, err := h.scores.Results(r.Context(), scoreFilterOf(r))
	if err != nil {
		writeError(w, r, err)
		return
	}
	response := raceResultsListResponse{Races: make([]raceResultsResponse, 0, len(races))}
	for _, race := range races {
		response.Races = append(response.Races, toRaceResultsResponse(race))
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *scoreHandlers) raceResults(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	race, err := h.scores.RaceResults(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toRaceResultsResponse(race))
}

func (h *scoreHandlers) standings(w http.ResponseWriter, r *http.Request) {
	standings, err := h.scores.Standings(r.Context(), scoreFilterOf(r))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toStandingsResponse(standings))
}

func scoreFilterOf(r *http.Request) scores.Filter {
	query := r.URL.Query()
	return scores.Filter{Season: query.Get("season"), Category: query.Get("category")}
}

func toRaceResponse(race scores.Race) raceResponse {
	return raceResponse(race)
}

func toRaceResultsResponse(race scores.RaceResults) raceResultsResponse {
	results := make([]driverResultResponse, 0, len(race.Results))
	for _, result := range race.Results {
		results = append(results, driverResultResponse(result))
	}
	return raceResultsResponse{Race: toRaceResponse(race.Race), Results: results}
}

func toStandingsResponse(standings scores.Standings) standingsResponse {
	response := standingsResponse{
		SeasonYear:   standings.SeasonYear,
		CategoryCode: standings.CategoryCode,
		Drivers:      make([]driverStandingResponse, 0, len(standings.Drivers)),
		Teams:        make([]teamStandingResponse, 0, len(standings.Teams)),
	}
	for _, driver := range standings.Drivers {
		response.Drivers = append(response.Drivers, driverStandingResponse(driver))
	}
	for _, team := range standings.Teams {
		response.Teams = append(response.Teams, teamStandingResponse(team))
	}
	return response
}
