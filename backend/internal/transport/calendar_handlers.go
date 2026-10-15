package transport

import (
	"context"
	"net/http"
	"time"

	"github.com/racecontrol/backend/internal/calendar"
)

type CalendarManager interface {
	List(ctx context.Context, filter calendar.Filter) ([]calendar.Event, error)
	Get(ctx context.Context, id int64) (calendar.Event, error)
	Circuits(ctx context.Context) ([]calendar.Circuit, error)
	Seasons(ctx context.Context) ([]calendar.Season, error)
	Create(ctx context.Context, input calendar.NewEvent) (calendar.Event, error)
	Update(ctx context.Context, id int64, changes calendar.EventChanges) (calendar.Event, error)
	Cancel(ctx context.Context, id int64) (calendar.Event, error)
}

type calendarHandlers struct {
	calendar CalendarManager
}

type eventResponse struct {
	ID           int64     `json:"id"`
	Kind         string    `json:"kind"`
	Name         string    `json:"name"`
	CategoryCode string    `json:"categoryCode"`
	CategoryName string    `json:"categoryName"`
	SeasonID     int64     `json:"seasonId"`
	SeasonYear   int32     `json:"seasonYear"`
	RoundNumber  *int32    `json:"roundNumber"`
	GPName       *string   `json:"gpName"`
	ControlType  *string   `json:"controlType"`
	CircuitID    *int64    `json:"circuitId"`
	CircuitName  *string   `json:"circuitName"`
	CountryCode  *string   `json:"countryCode"`
	StartsAt     time.Time `json:"startsAt"`
	EndsAt       time.Time `json:"endsAt"`
	Status       string    `json:"status"`
	IsActive     bool      `json:"isActive"`
}

type eventListResponse struct {
	Events []eventResponse `json:"events"`
}

type circuitResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Locality    string `json:"locality"`
	CountryCode string `json:"countryCode"`
}

type circuitListResponse struct {
	Circuits []circuitResponse `json:"circuits"`
}

type seasonResponse struct {
	ID   int64 `json:"id"`
	Year int32 `json:"year"`
}

type seasonListResponse struct {
	Seasons []seasonResponse `json:"seasons"`
}

type createEventRequest struct {
	Kind         string    `json:"kind"`
	StartsAt     time.Time `json:"startsAt"`
	EndsAt       time.Time `json:"endsAt"`
	SeasonID     int64     `json:"seasonId"`
	CategoryCode string    `json:"categoryCode"`
	CircuitID    *int64    `json:"circuitId"`
	Number       *int32    `json:"number"`
	GPName       *string   `json:"gpName"`
	ControlType  *string   `json:"controlType"`
}

type updateEventRequest struct {
	StartsAt     *time.Time `json:"startsAt"`
	EndsAt       *time.Time `json:"endsAt"`
	SeasonID     *int64     `json:"seasonId"`
	CategoryCode *string    `json:"categoryCode"`
	CircuitID    *int64     `json:"circuitId"`
	Number       *int32     `json:"number"`
	GPName       *string    `json:"gpName"`
	ControlType  *string    `json:"controlType"`
}

func (h *calendarHandlers) list(w http.ResponseWriter, r *http.Request) {
	events, err := h.calendar.List(r.Context(), filterOf(r))
	if err != nil {
		writeError(w, r, err)
		return
	}

	now := time.Now()
	response := eventListResponse{Events: make([]eventResponse, 0, len(events))}
	for _, event := range events {
		response.Events = append(response.Events, toEventResponse(event, now))
	}
	writeJSON(w, http.StatusOK, response)
}

// The downloads honour the same filters as the list, so the file matches what
// the calendar is showing.
func (h *calendarHandlers) downloadICS(w http.ResponseWriter, r *http.Request) {
	events, err := h.calendar.List(r.Context(), filterOf(r))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeDownload(w, r, "text/calendar; charset=utf-8", "calendario-racecontrol.ics", calendar.BuildICS(events, time.Now()))
}

func (h *calendarHandlers) downloadPDF(w http.ResponseWriter, r *http.Request) {
	events, err := h.calendar.List(r.Context(), filterOf(r))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeDownload(w, r, "application/pdf", "calendario-racecontrol.pdf", calendar.BuildPDF(events, time.Now()))
}

func filterOf(r *http.Request) calendar.Filter {
	query := r.URL.Query()
	return calendar.Filter{
		Category: query.Get("category"),
		Kind:     query.Get("kind"),
		Period:   query.Get("period"),
	}
}

func (h *calendarHandlers) get(w http.ResponseWriter, r *http.Request) {
	h.respondWithEvent(w, r, http.StatusOK, func(id int64) (calendar.Event, error) {
		return h.calendar.Get(r.Context(), id)
	})
}

func (h *calendarHandlers) create(w http.ResponseWriter, r *http.Request) {
	var request createEventRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, r, err)
		return
	}
	event, err := h.calendar.Create(r.Context(), calendar.NewEvent{
		Kind:         request.Kind,
		StartsAt:     request.StartsAt,
		EndsAt:       request.EndsAt,
		SeasonID:     request.SeasonID,
		CategoryCode: request.CategoryCode,
		CircuitID:    request.CircuitID,
		Number:       request.Number,
		GPName:       request.GPName,
		ControlType:  request.ControlType,
	})
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toEventResponse(event, time.Now()))
}

func (h *calendarHandlers) update(w http.ResponseWriter, r *http.Request) {
	var request updateEventRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, r, err)
		return
	}
	h.respondWithEvent(w, r, http.StatusOK, func(id int64) (calendar.Event, error) {
		return h.calendar.Update(r.Context(), id, calendar.EventChanges(request))
	})
}

func (h *calendarHandlers) cancel(w http.ResponseWriter, r *http.Request) {
	h.respondWithEvent(w, r, http.StatusOK, func(id int64) (calendar.Event, error) {
		return h.calendar.Cancel(r.Context(), id)
	})
}

func (h *calendarHandlers) circuits(w http.ResponseWriter, r *http.Request) {
	circuits, err := h.calendar.Circuits(r.Context())
	if err != nil {
		writeError(w, r, err)
		return
	}
	response := circuitListResponse{Circuits: make([]circuitResponse, 0, len(circuits))}
	for _, circuit := range circuits {
		response.Circuits = append(response.Circuits, circuitResponse(circuit))
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *calendarHandlers) seasons(w http.ResponseWriter, r *http.Request) {
	seasons, err := h.calendar.Seasons(r.Context())
	if err != nil {
		writeError(w, r, err)
		return
	}
	response := seasonListResponse{Seasons: make([]seasonResponse, 0, len(seasons))}
	for _, season := range seasons {
		response.Seasons = append(response.Seasons, seasonResponse{ID: season.ID, Year: season.Year})
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *calendarHandlers) respondWithEvent(w http.ResponseWriter, r *http.Request, status int,
	operation func(id int64) (calendar.Event, error)) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	event, err := operation(id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, status, toEventResponse(event, time.Now()))
}

func toEventResponse(event calendar.Event, now time.Time) eventResponse {
	return eventResponse{
		ID:           event.ID,
		Kind:         string(event.Kind),
		Name:         event.Name(),
		CategoryCode: event.CategoryCode,
		CategoryName: event.CategoryName,
		SeasonID:     event.SeasonID,
		SeasonYear:   event.SeasonYear,
		RoundNumber:  event.RoundNumber,
		GPName:       event.GPName,
		ControlType:  event.ControlType,
		CircuitID:    event.CircuitID,
		CircuitName:  event.CircuitName,
		CountryCode:  event.CountryCode,
		StartsAt:     event.StartsAt,
		EndsAt:       event.EndsAt,
		Status:       event.Status(now),
		IsActive:     event.IsActive,
	}
}
