package transport

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/racecontrol/backend/internal/auth"
	"github.com/racecontrol/backend/internal/domain"
)

const maxRequestBodyBytes = 64 << 10

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

type errorMapping struct {
	target     error
	status     int
	code       string
	message    string
	retryAfter time.Duration
}

var errorMappings = []errorMapping{
	{target: domain.ErrInvalidCredentials, status: http.StatusUnauthorized, code: "invalid_credentials", message: "Usuario o contraseña incorrectos"},
	{target: domain.ErrUnauthenticated, status: http.StatusUnauthorized, code: "unauthenticated", message: "Tenés que iniciar sesión"},
	{target: domain.ErrAccountLocked, status: http.StatusTooManyRequests, code: "account_locked", message: "Demasiados intentos fallidos. Probá de nuevo en 15 minutos", retryAfter: auth.LockoutWindow},
	{target: domain.ErrForbidden, status: http.StatusForbidden, code: "forbidden", message: "No tenés permiso para esta operación"},
	{target: domain.ErrNotFound, status: http.StatusNotFound, code: "not_found", message: "No encontrado"},
	{target: domain.ErrSelfLockout, status: http.StatusConflict, code: "self_lockout", message: "No podés desactivar tu propia cuenta ni cambiar tu propio rol"},
	{target: domain.ErrConflict, status: http.StatusConflict, code: "conflict", message: "El recurso ya existe"},
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	var validationErr *domain.ValidationError
	if errors.As(err, &validationErr) {
		writeJSON(w, http.StatusBadRequest, errorBody{Code: "validation", Message: validationErr.Message, Field: validationErr.Field})
		return
	}
	var conflictErr *domain.ConflictError
	if errors.As(err, &conflictErr) {
		writeJSON(w, http.StatusConflict, errorBody{Code: "conflict", Message: conflictErr.Message, Field: conflictErr.Field})
		return
	}

	for _, mapping := range errorMappings {
		if errors.Is(err, mapping.target) {
			if mapping.retryAfter > 0 {
				w.Header().Set("Retry-After", strconv.Itoa(int(mapping.retryAfter.Seconds())))
			}
			writeJSON(w, mapping.status, errorBody{Code: mapping.code, Message: mapping.message})
			return
		}
	}

	slog.ErrorContext(r.Context(), "unhandled request error", "path", r.URL.Path, "error", err)
	writeJSON(w, http.StatusInternalServerError, errorBody{Code: "internal", Message: "Error interno del servidor"})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("encode response body", "error", err)
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBodyBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return domain.NewValidationError("body", "el cuerpo de la petición no es un JSON válido")
	}
	return nil
}
