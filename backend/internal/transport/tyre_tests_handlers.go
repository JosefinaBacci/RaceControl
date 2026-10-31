package transport

import (
	"context"
	"net/http"

	"github.com/racecontrol/backend/internal/domain"
	"github.com/racecontrol/backend/internal/results"
	"github.com/racecontrol/backend/internal/tyretests"
)

type TyreTestRecorder interface {
	Types(ctx context.Context) ([]results.Type, error)
	History(ctx context.Context, filter domain.HistoryFilter) ([]tyretests.Result, error)
	Chain(ctx context.Context, eventID int64) ([]tyretests.Result, error)
	Record(ctx context.Context, eventID, actorUserID int64, input tyretests.NewResult) (tyretests.Result, error)
	Correct(ctx context.Context, resultID, actorUserID int64, input tyretests.Correction) (tyretests.Result, error)
}

type tyreTestHandlers struct {
	tyreTests TyreTestRecorder
}

// tyreTestResultResponse is the public row: the shared result plus the circuit
// the test ran on.
type tyreTestResultResponse struct {
	resultEntryResponse
	CircuitName string `json:"circuitName"`
}

type tyreTestResultListResponse struct {
	Results []tyreTestResultResponse `json:"results"`
}

// tyreTestRevisionResponse is the FIA row, with the author, the result it
// corrects and whether it is still the current one.
type tyreTestRevisionResponse struct {
	resultRevisionResponse
	CircuitName string `json:"circuitName"`
	IsCurrent   bool   `json:"isCurrent"`
}

type tyreTestRevisionListResponse struct {
	Results []tyreTestRevisionResponse `json:"results"`
}

type recordTyreTestResultRequest struct {
	resultInputRequest
	CompoundCode string `json:"compoundCode"`
}

func (h *tyreTestHandlers) types(w http.ResponseWriter, r *http.Request) {
	types, err := h.tyreTests.Types(r.Context())
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toHistoryTypeListResponse(types))
}

func (h *tyreTestHandlers) history(w http.ResponseWriter, r *http.Request) {
	filter, err := historyFilterOf(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	found, err := h.tyreTests.History(r.Context(), filter)
	if err != nil {
		writeError(w, r, err)
		return
	}
	response := tyreTestResultListResponse{Results: make([]tyreTestResultResponse, 0, len(found))}
	for _, result := range found {
		response.Results = append(response.Results, tyreTestResultResponse{
			resultEntryResponse: toResultEntryResponse(result.Entry),
			CircuitName:         result.CircuitName,
		})
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *tyreTestHandlers) chain(w http.ResponseWriter, r *http.Request) {
	eventID, err := pathID(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	found, err := h.tyreTests.Chain(r.Context(), eventID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	response := tyreTestRevisionListResponse{Results: make([]tyreTestRevisionResponse, 0, len(found))}
	for _, result := range found {
		response.Results = append(response.Results, toTyreTestRevisionResponse(result))
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *tyreTestHandlers) record(w http.ResponseWriter, r *http.Request) {
	var request recordTyreTestResultRequest
	h.respondWithWrite(w, r, &request, func(eventID, actorUserID int64) (tyretests.Result, error) {
		return h.tyreTests.Record(r.Context(), eventID, actorUserID, tyretests.NewResult{
			Input:        request.toInput(),
			CompoundCode: request.CompoundCode,
		})
	})
}

func (h *tyreTestHandlers) correct(w http.ResponseWriter, r *http.Request) {
	var request resultCorrectionRequest
	h.respondWithWrite(w, r, &request, func(resultID, actorUserID int64) (tyretests.Result, error) {
		return h.tyreTests.Correct(r.Context(), resultID, actorUserID, tyretests.Correction{
			Input:          request.toInput(),
			CorrectionNote: request.CorrectionNote,
		})
	})
}

// respondWithWrite runs what a record and a correction share: the path id, the
// acting administrator and the body, then answers 201 with the stored result.
func (h *tyreTestHandlers) respondWithWrite(w http.ResponseWriter, r *http.Request, request any, write func(id, actorUserID int64) (tyretests.Result, error)) {
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
	if err := decodeJSON(w, r, request); err != nil {
		writeError(w, r, err)
		return
	}
	result, err := write(id, actor.UserID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toTyreTestRevisionResponse(result))
}

func toTyreTestRevisionResponse(result tyretests.Result) tyreTestRevisionResponse {
	return tyreTestRevisionResponse{
		resultRevisionResponse: toResultRevisionResponse(result.Entry),
		CircuitName:            result.CircuitName,
		IsCurrent:              result.IsCurrent,
	}
}
