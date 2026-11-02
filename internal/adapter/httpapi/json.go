package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return jsonDecodeError(err, "invalid request body")
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return badRequest(
				"request body must contain exactly one JSON object",
			)
		}

		return jsonDecodeError(
			err,
			"request body must contain exactly one JSON object",
		)
	}
	return nil
}

func jsonDecodeError(err error, message string) error {
	if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
		return &HTTPError{
			Status:  http.StatusRequestEntityTooLarge,
			Message: "request body too large",
			Cause:   err,
		}
	}

	return &HTTPError{
		Status:  http.StatusBadRequest,
		Message: message,
		Cause:   err,
	}
}

func writeJSON(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	data any,
) {
	body, err := json.Marshal(data)
	if err != nil {
		writeHTTPError(
			w,
			r,
			fmt.Errorf("marshal JSON response: %w", err),
		)
		return
	}

	body = append(body, '\n')

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if _, err := w.Write(body); err != nil {
		slog.ErrorContext(
			r.Context(),
			"failed to write JSON response",
			"error",
			err,
		)
	}
}
