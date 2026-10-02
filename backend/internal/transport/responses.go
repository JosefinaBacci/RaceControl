package transport

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

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
	target  error
	status  int
	code    string
	message string
}

var errorMappings = []errorMapping{
	{domain.ErrInvalidCredentials, http.StatusUnauthorized, "invalid_credentials", "Usuario o contraseña incorrectos"},
	{domain.ErrUnauthenticated, http.StatusUnauthorized, "unauthenticated", "Tenés que iniciar sesión"},
	{domain.ErrAccountLocked, http.StatusTooManyRequests, "account_locked", "Demasiados intentos fallidos. Probá de nuevo en 15 minutos"},
	{domain.ErrForbidden, http.StatusForbidden, "forbidden", "No tenés permiso para esta operación"},
	{domain.ErrNotFound, http.StatusNotFound, "not_found", "No encontrado"},
	{domain.ErrConflict, http.StatusConflict, "conflict", "El recurso ya existe"},
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	var validationErr *domain.ValidationError
	if errors.As(err, &validationErr) {
		writeJSON(w, http.StatusBadRequest, errorBody{Code: "validation", Message: validationErr.Message, Field: validationErr.Field})
		return
	}

	for _, mapping := range errorMappings {
		if errors.Is(err, mapping.target) {
			if mapping.status == http.StatusTooManyRequests {
				w.Header().Set("Retry-After", strconv.Itoa(int(auth.LockoutWindow.Seconds())))
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
