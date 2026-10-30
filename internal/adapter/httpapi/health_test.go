package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	Health(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", recorder.Code, http.StatusOK)
	}
	if recorder.Body.String() != "OK" {
		t.Errorf("body = %q, want %q", recorder.Body.String(), "OK")
	}
	const expectedContentType = "text/plain; charset=utf-8"
	contentType := recorder.Header().Get("Content-Type")
	if contentType != expectedContentType {
		t.Errorf("content-type = %q, want %q", contentType, expectedContentType)
	}
}
