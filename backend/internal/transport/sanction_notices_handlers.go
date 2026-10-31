package transport

import (
	"context"
	"net/http"
	"time"

	"github.com/racecontrol/backend/internal/domain"
	"github.com/racecontrol/backend/internal/sanctions"
)

type SanctionNoticeManager interface {
	TeamNotices(ctx context.Context, teamID int64) (sanctions.NoticeBoard, error)
	Acknowledge(ctx context.Context, identity domain.Identity, versionID int64) (sanctions.Notice, error)
	FollowUps(ctx context.Context) ([]sanctions.FollowUp, error)
}

type sanctionNoticeHandlers struct {
	notices SanctionNoticeManager
}

type acknowledgementFields struct {
	Status         string     `json:"status"`
	AcknowledgedAt *time.Time `json:"acknowledgedAt"`
	AcknowledgedBy *string    `json:"acknowledgedBy"`
}

// sanctionNoticeResponse is identified by the version: that is what the team
// confirms.
type sanctionNoticeResponse struct {
	ID        int64            `json:"id"`
	Sanction  sanctionResponse `json:"sanction"`
	IsOwnTeam bool             `json:"isOwnTeam"`
	acknowledgementFields
}

type sanctionNoticeBoardResponse struct {
	PendingCount int                      `json:"pendingCount"`
	Notices      []sanctionNoticeResponse `json:"notices"`
}

type recipientResponse struct {
	Team sanctionTeamResponse `json:"team"`
	acknowledgementFields
}

type versionFollowUpResponse struct {
	sanctionVersionResponse
	ChangedBy      string              `json:"changedBy"`
	IsCurrent      bool                `json:"isCurrent"`
	ConfirmedCount int                 `json:"confirmedCount"`
	Teams          []recipientResponse `json:"teams"`
}

type followUpResponse struct {
	Sanction sanctionResponse          `json:"sanction"`
	Versions []versionFollowUpResponse `json:"versions"`
}

type followUpListResponse struct {
	FollowUps []followUpResponse `json:"followUps"`
}

func (h *sanctionNoticeHandlers) teamNotices(w http.ResponseWriter, r *http.Request) {
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
	response := sanctionNoticeBoardResponse{PendingCount: board.PendingCount, Notices: make([]sanctionNoticeResponse, 0, len(board.Notices))}
	for _, notice := range board.Notices {
		response.Notices = append(response.Notices, toSanctionNoticeResponse(notice))
	}
	writeJSON(w, http.StatusOK, response)
}

// acknowledge answers the notice as it stands afterwards. Confirming twice is not
// an error: the first confirmation is kept and returned again.
func (h *sanctionNoticeHandlers) acknowledge(w http.ResponseWriter, r *http.Request) {
	identity, err := identityFrom(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	versionID, err := pathID(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	notice, err := h.notices.Acknowledge(r.Context(), identity, versionID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toSanctionNoticeResponse(notice))
}

func (h *sanctionNoticeHandlers) followUps(w http.ResponseWriter, r *http.Request) {
	followUps, err := h.notices.FollowUps(r.Context())
	if err != nil {
		writeError(w, r, err)
		return
	}
	response := followUpListResponse{FollowUps: make([]followUpResponse, 0, len(followUps))}
	for _, followUp := range followUps {
		response.FollowUps = append(response.FollowUps, toFollowUpResponse(followUp))
	}
	writeJSON(w, http.StatusOK, response)
}

func toSanctionNoticeResponse(notice sanctions.Notice) sanctionNoticeResponse {
	return sanctionNoticeResponse{
		ID:                    notice.Sanction.Current.ID,
		Sanction:              toSanctionResponse(notice.Sanction),
		IsOwnTeam:             notice.IsOwnTeam,
		acknowledgementFields: toAcknowledgementFields(notice.Acknowledgement),
	}
}

func toFollowUpResponse(followUp sanctions.FollowUp) followUpResponse {
	versions := make([]versionFollowUpResponse, 0, len(followUp.Versions))
	for _, version := range followUp.Versions {
		versions = append(versions, toVersionFollowUpResponse(version))
	}
	return followUpResponse{Sanction: toSanctionResponse(followUp.Sanction), Versions: versions}
}

func toVersionFollowUpResponse(version sanctions.VersionFollowUp) versionFollowUpResponse {
	teams := make([]recipientResponse, 0, len(version.Teams))
	confirmed := 0
	for _, team := range version.Teams {
		if team.Acknowledgement.Status == sanctions.AcknowledgementConfirmed {
			confirmed++
		}
		teams = append(teams, recipientResponse{
			Team:                  sanctionTeamResponse(team.Team),
			acknowledgementFields: toAcknowledgementFields(team.Acknowledgement),
		})
	}
	return versionFollowUpResponse{
		sanctionVersionResponse: toSanctionVersionResponse(version.Version),
		ChangedBy:               version.ChangedBy,
		IsCurrent:               version.IsCurrent,
		ConfirmedCount:          confirmed,
		Teams:                   teams,
	}
}

func toAcknowledgementFields(acknowledgement sanctions.Acknowledgement) acknowledgementFields {
	return acknowledgementFields{
		Status:         string(acknowledgement.Status),
		AcknowledgedAt: acknowledgement.At,
		AcknowledgedBy: acknowledgement.By,
	}
}
