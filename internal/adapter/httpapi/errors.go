package httpapi

import (
	"errors"
	"log/slog"
	"net/http"
)

type HTTPError struct {
	Status  int
	Message string
	Cause   error
}

func (e *HTTPError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *HTTPError) Unwrap() error {
	return e.Cause
}

func badRequest(message string) error {
	return &HTTPError{
		Status:  http.StatusBadRequest,
		Message: message,
	}
}

func writeHTTPError(w http.ResponseWriter, r *http.Request, err error) {
	if httpErr, ok := errors.AsType[*HTTPError](err); ok {
		http.Error(w, httpErr.Message, httpErr.Status)
		return
	}

	slog.ErrorContext(r.Context(), "HTTP request failed", "error", err)

	http.Error(w, "internal server error", http.StatusInternalServerError)
}
