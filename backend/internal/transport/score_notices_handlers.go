package transport

import (
	"context"
	"net/http"
	"time"

	"github.com/racecontrol/backend/internal/domain"
	"github.com/racecontrol/backend/internal/scores"
)

type ScoreNoticeManager interface {
	TeamNotices(ctx context.Context, teamID int64) (scores.NoticeBoard, error)
	Acknowledge(ctx context.Context, identity domain.Identity, changeLogID int64) (scores.Notice, error)
	Changes(ctx context.Context) ([]scores.ScoreChange, error)
}

type scoreNoticeHandlers struct {
	notices ScoreNoticeManager
}

type changedPointsResponse struct {
	DriverID  int64    `json:"driverId"`
	FirstName string   `json:"firstName"`
	LastName  string   `json:"lastName"`
	OldPoints *float64 `json:"oldPoints"`
	NewPoints *float64 `json:"newPoints"`
}

type noticeResponse struct {
	ID             int64                   `json:"id"`
	Race           raceResponse            `json:"race"`
	Kind           string                  `json:"kind"`
	ChangedAt      time.Time               `json:"changedAt"`
	Status         string                  `json:"status"`
	AcknowledgedAt *time.Time              `json:"acknowledgedAt"`
	AcknowledgedBy *string                 `json:"acknowledgedBy"`
	Items          []changedPointsResponse `json:"items"`
}

type noticeBoardResponse struct {
	PendingCount int              `json:"pendingCount"`
	Notices      []noticeResponse `json:"notices"`
}

type teamAcknowledgementResponse struct {
	TeamID         int64      `json:"teamId"`
	TeamCode       string     `json:"teamCode"`
	TeamName       string     `json:"teamName"`
	Status         string     `json:"status"`
	AcknowledgedAt *time.Time `json:"acknowledgedAt"`
	AcknowledgedBy *string    `json:"acknowledgedBy"`
}

type scoreChangeResponse struct {
	ID        int64                         `json:"id"`
	Race      raceResponse                  `json:"race"`
	Kind      string                        `json:"kind"`
	ChangedAt time.Time                     `json:"changedAt"`
	ChangedBy string                        `json:"changedBy"`
	Teams     []teamAcknowledgementResponse `json:"teams"`
}

type scoreChangeListResponse struct {
	Changes []scoreChangeResponse `json:"changes"`
}

func (h *scoreNoticeHandlers) teamNotices(w http.ResponseWriter, r *http.Request) {
	teamID, err := teamIDFrom(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	board, err := h.notices.TeamNotices(r.Context(), teamID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	response := noticeBoardResponse{PendingCount: board.PendingCount, Notices: make([]noticeResponse, 0, len(board.Notices))}
	for _, notice := range board.Notices {
		response.Notices = append(response.Notices, toNoticeResponse(notice))
	}
	writeJSON(w, http.StatusOK, response)
}

// acknowledge answers the notice as it stands afterwards. Confirming twice is not
// an error: the first confirmation is kept and returned again.
func (h *scoreNoticeHandlers) acknowledge(w http.ResponseWriter, r *http.Request) {
	identity, err := identityFrom(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	id, err := pathID(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	notice, err := h.notices.Acknowledge(r.Context(), identity, id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toNoticeResponse(notice))
}

func (h *scoreNoticeHandlers) changes(w http.ResponseWriter, r *http.Request) {
	changes, err := h.notices.Changes(r.Context())
	if err != nil {
		writeError(w, r, err)
		return
	}
	response := scoreChangeListResponse{Changes: make([]scoreChangeResponse, 0, len(changes))}
	for _, change := range changes {
		response.Changes = append(response.Changes, toScoreChangeResponse(change))
	}
	writeJSON(w, http.StatusOK, response)
}

func toChangedPointsResponses(changed []scores.ChangedPoints) []changedPointsResponse {
	items := make([]changedPointsResponse, 0, len(changed))
	for _, item := range changed {
		items = append(items, changedPointsResponse(item))
	}
	return items
}

func toNoticeResponse(notice scores.Notice) noticeResponse {
	return noticeResponse{
		ID:             notice.ChangeLogID,
		Race:           toRaceResponse(notice.Race),
		Kind:           string(notice.Kind),
		ChangedAt:      notice.ChangedAt,
		Status:         string(notice.Acknowledgement.Status),
		AcknowledgedAt: notice.Acknowledgement.At,
		AcknowledgedBy: notice.Acknowledgement.By,
		Items:          toChangedPointsResponses(notice.Items),
	}
}

func toScoreChangeResponse(change scores.ScoreChange) scoreChangeResponse {
	teams := make([]teamAcknowledgementResponse, 0, len(change.Teams))
	for _, team := range change.Teams {
		teams = append(teams, teamAcknowledgementResponse{
			TeamID:         team.TeamID,
			TeamCode:       team.TeamCode,
			TeamName:       team.TeamName,
			Status:         string(team.Acknowledgement.Status),
			AcknowledgedAt: team.Acknowledgement.At,
			AcknowledgedBy: team.Acknowledgement.By,
		})
	}
	return scoreChangeResponse{
		ID:        change.ChangeLogID,
		Race:      toRaceResponse(change.Race),
		Kind:      string(change.Kind),
		ChangedAt: change.ChangedAt,
		ChangedBy: change.ChangedBy,
		Teams:     teams,
	}
}
