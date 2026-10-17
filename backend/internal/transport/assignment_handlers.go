package transport

import (
	"github.com/racecontrol/backend/internal/domain"
	"net/http"
	"strconv"
	"time"
)

type assignmentResponse struct {
	ID               int64      `json:"id"`
	DriverID         int64      `json:"driverId"`
	FirstName        string     `json:"firstName"`
	LastName         string     `json:"lastName"`
	TeamID           *int64     `json:"teamId"`
	TeamName         string     `json:"teamName"`
	PreviousTeamName string     `json:"previousTeamName"`
	IsStarter        bool       `json:"isStarter"`
	EffectiveFrom    time.Time  `json:"effectiveFrom"`
	EffectiveUntil   *time.Time `json:"effectiveUntil"`
}

func (h *driverHandlers) history(w http.ResponseWriter, r *http.Request) {
	identity, err := identityFrom(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	var teamID *int64
	if raw := r.URL.Query().Get("teamId"); raw != "" {
		value, parseErr := strconv.ParseInt(raw, 10, 64)
		if parseErr != nil || value <= 0 {
			writeError(w, r, domain.NewValidationError("teamId", "la escudería no es válida"))
			return
		}
		teamID = &value
	}
	history, err := h.drivers.History(r.Context(), identity, teamID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	rows := make([]assignmentResponse, 0, len(history))
	for _, row := range history {
		rows = append(rows, assignmentResponse(row))
	}
	writeJSON(w, http.StatusOK, struct {
		Assignments []assignmentResponse `json:"assignments"`
	}{rows})
}
