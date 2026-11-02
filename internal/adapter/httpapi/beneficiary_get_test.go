package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Imanghvs/crudgo/internal/beneficiary"
	"github.com/google/uuid"
)

func TestBeneficiaryHandler_GetByID(t *testing.T) {
	expected := beneficiary.Beneficiary{
		ID:         uuid.New(),
		MerchantID: uuid.New(),
		Name:       "TestName",
		IBAN:       "TEST-IBAN-001",
		CreatedAt:  time.Now(),
	}

	repo := &fakeRepo{
		beneficiaries: []beneficiary.Beneficiary{expected},
	}
	handler := NewBeneficiaryHandler(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/beneficiaries/"+expected.ID.String(),
		nil,
	)
	req.SetPathValue("id", expected.ID.String())

	recorder := httptest.NewRecorder()
	handler.GetBeneficiaryByID(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"status = %d, want %d",
			recorder.Code,
			http.StatusOK,
		)
	}

	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("content type = %q, want application/json", got)
	}

	if repo.getByIDCalls != 1 {
		t.Errorf("repository calls = %d, want 1", repo.getByIDCalls)
	}

	var response BeneficiaryResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.ID != expected.ID {
		t.Errorf("id = %s, want %s", response.ID, expected.ID)
	}

	if response.MerchantID != expected.MerchantID {
		t.Errorf(
			"merchant id = %s, want %s",
			response.MerchantID,
			expected.MerchantID,
		)
	}

	if response.Name != expected.Name {
		t.Errorf("name = %q, want %q", response.Name, expected.Name)
	}

	if response.IBAN != expected.IBAN {
		t.Errorf("iban = %q, want %q", response.IBAN, expected.IBAN)
	}

	if !response.CreatedAt.Equal(expected.CreatedAt) {
		t.Errorf(
			"created_at = %v, want %v",
			response.CreatedAt,
			expected.CreatedAt,
		)
	}
}

func TestBeneficiaryHandler_GetByID_InvalidID(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{
			name: "malformed UUID",
			id:   "not-a-uuid",
		},
		{
			name: "nil UUID",
			id:   uuid.Nil.String(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRepo{}
			handler := NewBeneficiaryHandler(repo)

			req := httptest.NewRequest(
				http.MethodGet,
				"/beneficiaries/"+tc.id,
				nil,
			)
			req.SetPathValue("id", tc.id)

			recorder := httptest.NewRecorder()
			handler.GetBeneficiaryByID(recorder, req)

			if recorder.Code != http.StatusBadRequest {
				t.Errorf(
					"status = %d, want %d",
					recorder.Code,
					http.StatusBadRequest,
				)
			}

			if got := recorder.Body.String(); got != "invalid beneficiary id\n" {
				t.Errorf(
					"body = %q, want %q",
					got,
					"invalid beneficiary id\n",
				)
			}

			if repo.getByIDCalls != 0 {
				t.Errorf(
					"repository calls = %d, want 0",
					repo.getByIDCalls,
				)
			}
		})
	}
}

func TestBeneficiaryHandler_GetByID_RepositoryError(t *testing.T) {
	tests := []struct {
		name       string
		repoErr    error
		wantStatus int
		wantBody   string
	}{
		{
			name: "beneficiary not found",
			repoErr: fmt.Errorf(
				"get beneficiary by id: %w",
				beneficiary.ErrNotFound,
			),
			wantStatus: http.StatusNotFound,
			wantBody:   "beneficiary not found\n",
		},
		{
			name:       "unexpected database error",
			repoErr:    errors.New("database connection failed: internal details"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "internal server error\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRepo{
				getByIDErr: tc.repoErr,
			}
			handler := NewBeneficiaryHandler(repo)

			id := uuid.New()

			req := httptest.NewRequest(
				http.MethodGet,
				"/beneficiaries/"+id.String(),
				nil,
			)
			req.SetPathValue("id", id.String())

			recorder := httptest.NewRecorder()
			handler.GetBeneficiaryByID(recorder, req)

			if recorder.Code != tc.wantStatus {
				t.Errorf(
					"status = %d, want %d",
					recorder.Code,
					tc.wantStatus,
				)
			}

			if got := recorder.Body.String(); got != tc.wantBody {
				t.Errorf(
					"body = %q, want %q",
					got,
					tc.wantBody,
				)
			}

			if repo.getByIDCalls != 1 {
				t.Errorf(
					"repository calls = %d, want 1",
					repo.getByIDCalls,
				)
			}
		})
	}
}
