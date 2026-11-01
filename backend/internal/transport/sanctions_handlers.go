package transport

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/racecontrol/backend/internal/domain"
	"github.com/racecontrol/backend/internal/sanctions"
)

type SanctionRegistry interface {
	PenaltyTypes(ctx context.Context) ([]sanctions.PenaltyType, error)
	List(ctx context.Context, filter sanctions.Filter) ([]sanctions.Sanction, error)
	Get(ctx context.Context, id int64) (sanctions.Detail, error)
	Issue(ctx context.Context, actorUserID int64, input sanctions.NewSanction) (sanctions.Detail, error)
	Amend(ctx context.Context, actorUserID, sanctionID int64, amendment sanctions.Amendment) (sanctions.Detail, error)
	Revoke(ctx context.Context, actorUserID, sanctionID int64, note string) (sanctions.Detail, error)
}

type sanctionHandlers struct {
	sanctions SanctionRegistry
}

type penaltyTypeResponse struct {
	Code     string  `json:"code"`
	Name     string  `json:"name"`
	Unit     *string `json:"unit"`
	MaxValue *int32  `json:"maxValue"`
}

type penaltyTypeListResponse struct {
	PenaltyTypes []penaltyTypeResponse `json:"penaltyTypes"`
}

type penaltyResponse struct {
	Code  string  `json:"code"`
	Name  string  `json:"name"`
	Unit  *string `json:"unit"`
	Value *int32  `json:"value"`
}

type sanctionSubjectResponse struct {
	Kind            string  `json:"kind"`
	DriverID        *int64  `json:"driverId"`
	DriverFirstName *string `json:"driverFirstName"`
	DriverLastName  *string `json:"driverLastName"`
}

type sanctionTeamResponse struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type sanctionEventResponse struct {
	ID       int64     `json:"id"`
	Name     string    `json:"name"`
	StartsAt time.Time `json:"startsAt"`
}

// sanctionVersionResponse carries no author: the public record speaks for the
// FIA, and usernames stay in the FIA-only follow-up.
type sanctionVersionResponse struct {
	ID        int64                  `json:"id"`
	Action    string                 `json:"action"`
	Penalty   penaltyResponse        `json:"penalty"`
	Reason    string                 `json:"reason"`
	Event     *sanctionEventResponse `json:"event"`
	Note      *string                `json:"note"`
	ChangedAt time.Time              `json:"changedAt"`
}

type sanctionResponse struct {
	ID           int64                   `json:"id"`
	Status       string                  `json:"status"`
	Subject      sanctionSubjectResponse `json:"subject"`
	Team         sanctionTeamResponse    `json:"team"`
	CategoryCode string                  `json:"categoryCode"`
	CategoryName string                  `json:"categoryName"`
	IssuedAt     time.Time               `json:"issuedAt"`
	Current      sanctionVersionResponse `json:"current"`
}

type sanctionListResponse struct {
	Sanctions []sanctionResponse `json:"sanctions"`
}

type sanctionDetailResponse struct {
	sanctionResponse
	Versions []sanctionVersionResponse `json:"versions"`
}

type sanctionTermsRequest struct {
	PenaltyType  string `json:"penaltyType"`
	PenaltyValue *int32 `json:"penaltyValue"`
	Reason       string `json:"reason"`
	EventID      *int64 `json:"eventId"`
}

type issueSanctionRequest struct {
	DriverID *int64 `json:"driverId"`
	TeamID   *int64 `json:"teamId"`
	sanctionTermsRequest
}

type amendSanctionRequest struct {
	sanctionTermsRequest
	Note string `json:"note"`
}

type revokeSanctionRequest struct {
	Note string `json:"note"`
}

func (h *sanctionHandlers) penaltyTypes(w http.ResponseWriter, r *http.Request) {
	types, err := h.sanctions.PenaltyTypes(r.Context())
	if err != nil {
		writeError(w, r, err)
		return
	}
	response := penaltyTypeListResponse{PenaltyTypes: make([]penaltyTypeResponse, 0, len(types))}
	for _, penaltyType := range types {
		response.PenaltyTypes = append(response.PenaltyTypes, penaltyTypeResponse{
			Code: penaltyType.Code, Name: penaltyType.Name, Unit: penaltyType.Unit, MaxValue: penaltyType.MaxValue,
		})
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *sanctionHandlers) list(w http.ResponseWriter, r *http.Request) {
	filter, err := sanctionFilterOf(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	list, err := h.sanctions.List(r.Context(), filter)
	if err != nil {
		writeError(w, r, err)
		return
	}
	response := sanctionListResponse{Sanctions: make([]sanctionResponse, 0, len(list))}
	for _, sanction := range list {
		response.Sanctions = append(response.Sanctions, toSanctionResponse(sanction))
	}
	writeJSON(w, http.StatusOK, response)
}

func sanctionFilterOf(r *http.Request) (sanctions.Filter, error) {
	query := r.URL.Query()
	dateRange, err := dateRangeOf(r)
	if err != nil {
		return sanctions.Filter{}, err
	}
	driverID, err := optionalQueryID(query.Get("driverId"), "driverId")
	if err != nil {
		return sanctions.Filter{}, err
	}
	teamID, err := optionalQueryID(query.Get("teamId"), "teamId")
	if err != nil {
		return sanctions.Filter{}, err
	}
	return sanctions.Filter{
		DriverID: driverID,
		TeamID:   teamID,
		Category: query.Get("category"),
		Status:   sanctions.Status(strings.TrimSpace(query.Get("status"))),
		Range:    dateRange,
	}, nil
}

func optionalQueryID(raw, field string) (*int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id < 1 {
		return nil, domain.NewValidationError(field, "el identificador debe ser un entero positivo")
	}
	return &id, nil
}

func (h *sanctionHandlers) get(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	detail, err := h.sanctions.Get(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toSanctionDetailResponse(detail))
}

func (h *sanctionHandlers) issue(w http.ResponseWriter, r *http.Request) {
	var request issueSanctionRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, r, err)
		return
	}
	h.write(w, r, http.StatusCreated, func(identity domain.Identity) (sanctions.Detail, error) {
		return h.sanctions.Issue(r.Context(), identity.UserID, sanctions.NewSanction{
			Target: sanctions.Target{DriverID: request.DriverID, TeamID: request.TeamID},
			Terms:  sanctions.Terms(request.sanctionTermsRequest),
		})
	})
}

func (h *sanctionHandlers) amend(w http.ResponseWriter, r *http.Request) {
	var request amendSanctionRequest
	h.writeOnSanction(w, r, &request, func(identity domain.Identity, id int64) (sanctions.Detail, error) {
		return h.sanctions.Amend(r.Context(), identity.UserID, id, sanctions.Amendment{
			Terms: sanctions.Terms(request.sanctionTermsRequest),
			Note:  request.Note,
		})
	})
}

func (h *sanctionHandlers) revoke(w http.ResponseWriter, r *http.Request) {
	var request revokeSanctionRequest
	h.writeOnSanction(w, r, &request, func(identity domain.Identity, id int64) (sanctions.Detail, error) {
		return h.sanctions.Revoke(r.Context(), identity.UserID, id, request.Note)
	})
}

// writeOnSanction resolves the sanction of the path before reading the body, so
// a malformed id is a 404 whatever the body says.
func (h *sanctionHandlers) writeOnSanction(w http.ResponseWriter, r *http.Request, request any,
	change func(identity domain.Identity, id int64) (sanctions.Detail, error)) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if err := decodeJSON(w, r, request); err != nil {
		writeError(w, r, err)
		return
	}
	h.write(w, r, http.StatusOK, func(identity domain.Identity) (sanctions.Detail, error) {
		return change(identity, id)
	})
}

func (h *sanctionHandlers) write(w http.ResponseWriter, r *http.Request, status int,
	change func(identity domain.Identity) (sanctions.Detail, error)) {
	identity, err := identityFrom(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	detail, err := change(identity)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, status, toSanctionDetailResponse(detail))
}

func toSanctionResponse(sanction sanctions.Sanction) sanctionResponse {
	return sanctionResponse{
		ID:     sanction.ID,
		Status: sanctionStatusOf(sanction),
		Subject: sanctionSubjectResponse{
			Kind:            string(sanction.Subject.Kind),
			DriverID:        sanction.Subject.DriverID,
			DriverFirstName: sanction.Subject.DriverFirstName,
			DriverLastName:  sanction.Subject.DriverLastName,
		},
		Team:         sanctionTeamResponse(sanction.Team),
		CategoryCode: sanction.CategoryCode,
		CategoryName: sanction.CategoryName,
		IssuedAt:     sanction.IssuedAt.UTC(),
		Current:      toSanctionVersionResponse(sanction.Current),
	}
}

func sanctionStatusOf(sanction sanctions.Sanction) string {
	if sanction.IsRevoked() {
		return string(sanctions.StatusRevoked)
	}
	return string(sanctions.StatusActive)
}

func toSanctionDetailResponse(detail sanctions.Detail) sanctionDetailResponse {
	versions := make([]sanctionVersionResponse, 0, len(detail.Versions))
	for _, version := range detail.Versions {
		versions = append(versions, toSanctionVersionResponse(version))
	}
	return sanctionDetailResponse{sanctionResponse: toSanctionResponse(detail.Sanction), Versions: versions}
}

func toSanctionVersionResponse(version sanctions.Version) sanctionVersionResponse {
	return sanctionVersionResponse{
		ID:        version.ID,
		Action:    string(version.Action),
		Penalty:   penaltyResponse(version.Penalty),
		Reason:    version.Reason,
		Event:     toSanctionEventResponse(version.Event),
		Note:      version.Note,
		ChangedAt: version.ChangedAt.UTC(),
	}
}

func toSanctionEventResponse(event *sanctions.EventRef) *sanctionEventResponse {
	if event == nil {
		return nil
	}
	return &sanctionEventResponse{ID: event.ID, Name: event.Name, StartsAt: event.StartsAt.UTC()}
}
