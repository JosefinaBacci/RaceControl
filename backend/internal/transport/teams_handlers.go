package transport

import (
	"context"
	"net/http"

	"github.com/racecontrol/backend/internal/teams"
)

type TeamCatalog interface {
	List(ctx context.Context) ([]teams.Team, error)
}

type teamResponse struct {
	ID           int64  `json:"id"`
	Code         string `json:"code"`
	Name         string `json:"name"`
	CategoryCode string `json:"categoryCode"`
	CategoryName string `json:"categoryName"`
}

type teamListResponse struct {
	Teams []teamResponse `json:"teams"`
}

func listTeamsHandler(catalog TeamCatalog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list, err := catalog.List(r.Context())
		if err != nil {
			writeError(w, r, err)
			return
		}
		response := teamListResponse{Teams: make([]teamResponse, 0, len(list))}
		for _, team := range list {
			response.Teams = append(response.Teams, teamResponse(team))
		}
		writeJSON(w, http.StatusOK, response)
	}
}
