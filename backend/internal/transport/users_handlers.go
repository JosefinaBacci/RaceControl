package transport

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/racecontrol/backend/internal/domain"
	"github.com/racecontrol/backend/internal/users"
)

type UserManager interface {
	List(ctx context.Context, filter users.Filter) ([]users.Account, error)
	Get(ctx context.Context, id int64) (users.Account, error)
	Create(ctx context.Context, input users.NewAccount) (users.Account, error)
	Update(ctx context.Context, actor domain.Identity, id int64, changes users.AccountChanges) (users.Account, error)
	Deactivate(ctx context.Context, actor domain.Identity, id int64) (users.Account, error)
	Reactivate(ctx context.Context, id int64) (users.Account, error)
}

type userHandlers struct {
	users UserManager
}

type accountResponse struct {
	ID            int64      `json:"id"`
	Username      string     `json:"username"`
	Email         *string    `json:"email"`
	Role          string     `json:"role"`
	TeamID        *int64     `json:"teamId"`
	TeamName      *string    `json:"teamName"`
	IsActive      bool       `json:"isActive"`
	DeactivatedAt *time.Time `json:"deactivatedAt"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

type accountListResponse struct {
	Users []accountResponse `json:"users"`
}

type createAccountRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
	TeamID   *int64 `json:"teamId"`
}

type updateAccountRequest struct {
	Email    *string `json:"email"`
	Role     *string `json:"role"`
	TeamID   *int64  `json:"teamId"`
	Password *string `json:"password"`
}

func (h *userHandlers) list(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	accounts, err := h.users.List(r.Context(), users.Filter{
		Search: query.Get("q"),
		Role:   query.Get("role"),
		Status: query.Get("status"),
	})
	if err != nil {
		writeError(w, r, err)
		return
	}
	response := accountListResponse{Users: make([]accountResponse, 0, len(accounts))}
	for _, account := range accounts {
		response.Users = append(response.Users, toAccountResponse(account))
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *userHandlers) get(w http.ResponseWriter, r *http.Request) {
	h.respondWithAccount(w, r, http.StatusOK, func(id int64, _ domain.Identity) (users.Account, error) {
		return h.users.Get(r.Context(), id)
	})
}

func (h *userHandlers) create(w http.ResponseWriter, r *http.Request) {
	var request createAccountRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, r, err)
		return
	}
	account, err := h.users.Create(r.Context(), users.NewAccount(request))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toAccountResponse(account))
}

func (h *userHandlers) update(w http.ResponseWriter, r *http.Request) {
	var request updateAccountRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, r, err)
		return
	}
	h.respondWithAccount(w, r, http.StatusOK, func(id int64, actor domain.Identity) (users.Account, error) {
		return h.users.Update(r.Context(), actor, id, users.AccountChanges(request))
	})
}

func (h *userHandlers) deactivate(w http.ResponseWriter, r *http.Request) {
	h.respondWithAccount(w, r, http.StatusOK, func(id int64, actor domain.Identity) (users.Account, error) {
		return h.users.Deactivate(r.Context(), actor, id)
	})
}

func (h *userHandlers) reactivate(w http.ResponseWriter, r *http.Request) {
	h.respondWithAccount(w, r, http.StatusOK, func(id int64, _ domain.Identity) (users.Account, error) {
		return h.users.Reactivate(r.Context(), id)
	})
}

func (h *userHandlers) respondWithAccount(w http.ResponseWriter, r *http.Request, status int,
	operation func(id int64, actor domain.Identity) (users.Account, error)) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, r, domain.ErrNotFound)
		return
	}
	actor, err := identityFrom(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	account, err := operation(id, actor)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, status, toAccountResponse(account))
}

func toAccountResponse(account users.Account) accountResponse {
	return accountResponse{
		ID:            account.ID,
		Username:      account.Username,
		Email:         account.Email,
		Role:          string(account.Role),
		TeamID:        account.TeamID,
		TeamName:      account.TeamName,
		IsActive:      account.IsActive,
		DeactivatedAt: account.DeactivatedAt,
		CreatedAt:     account.CreatedAt,
		UpdatedAt:     account.UpdatedAt,
	}
}
