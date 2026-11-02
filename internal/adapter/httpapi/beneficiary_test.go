package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Imanghvs/crudgo/internal/beneficiary"
	"github.com/google/uuid"
)

type fakeRepo struct {
	beneficiaries []beneficiary.Beneficiary
	createErr     error
	createCalls   int
}

func (r *fakeRepo) Create(
	ctx context.Context,
	input beneficiary.CreateInput,
) (beneficiary.Beneficiary, error) {
	r.createCalls++

	if r.createErr != nil {
		return beneficiary.Beneficiary{}, r.createErr
	}

	created := beneficiary.Beneficiary{
		ID:         uuid.New(),
		MerchantID: input.MerchantID,
		Name:       input.Name,
		IBAN:       input.IBAN,
		CreatedAt:  time.Now(),
	}
	r.beneficiaries = append(r.beneficiaries, created)
	return created, nil
}

func TestBeneficiaryHandler_Create(t *testing.T) {
	repo := &fakeRepo{}
	handler := NewBeneficiaryHandler(repo)
	recorder := httptest.NewRecorder()
	input := CreateBeneficiaryRequest{
		MerchantID: uuid.New(),
		Name:       "TestName",
		IBAN:       "TEST-IBAN-001",
	}
	body, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("failed to marshal test request body: %v", err)
	}
	req := httptest.NewRequest(
		http.MethodPost,
		"/beneficiaries",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	handler.CreateBeneficiary(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}
	expectedContentType := "application/json"

	if contentType := recorder.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Fatalf("content type = %q, want %q", contentType, expectedContentType)
	}

	if len(repo.beneficiaries) != 1 {
		t.Fatalf(
			"created beneficiaries: %d, want 1",
			len(repo.beneficiaries),
		)
	}
	stored := repo.beneficiaries[0]

	var response CreateBeneficiaryResponse

	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.ID != stored.ID {
		t.Errorf("response ID = %s, want %s", response.ID, stored.ID)
	}

	if response.MerchantID != stored.MerchantID {
		t.Errorf("response merchant ID = %s, want %s", response.MerchantID, stored.MerchantID)
	}

	if response.Name != stored.Name {
		t.Errorf("response name = %q, want %q", response.Name, stored.Name)
	}

	if response.IBAN != stored.IBAN {
		t.Errorf("response IBAN = %q, want %q", response.IBAN, stored.IBAN)
	}

	if !response.CreatedAt.Equal(stored.CreatedAt) {
		t.Errorf("response created_at = %v, want %v", response.CreatedAt, stored.CreatedAt)
	}

	if stored.MerchantID != input.MerchantID {
		t.Errorf("repository merchant ID = %s, want %s", stored.MerchantID, input.MerchantID)
	}

	if stored.Name != input.Name {
		t.Errorf("repository name = %q, want %q", stored.Name, input.Name)
	}

	if stored.IBAN != input.IBAN {
		t.Errorf("repository IBAN = %q, want %q", stored.IBAN, input.IBAN)
	}
}

func TestBeneficiaryHandler_Create_InvalidRequest(t *testing.T) {
	merchantID := uuid.New()

	tests := []struct {
		name string
		body string
	}{
		{
			name: "malformed JSON",
			body: `{"merchant_id":`,
		},
		{
			name: "missing name",
			body: fmt.Sprintf(
				`{"merchant_id":%q,"iban":"TEST-IBAN-001"}`,
				merchantID,
			),
		},
		{
			name: "blank name",
			body: fmt.Sprintf(
				`{"merchant_id":%q,"name":"   ","iban":"TEST-IBAN-001"}`,
				merchantID,
			),
		},
		{
			name: "blank IBAN",
			body: fmt.Sprintf(
				`{"merchant_id":%q,"name":"TestName","iban":"   "}`,
				merchantID,
			),
		},
		{
			name: "missing merchant ID",
			body: `{"name":"TestName","iban":"TEST-IBAN-001"}`,
		},
		{
			name: "nil merchant ID",
			body: `{"merchant_id":"00000000-0000-0000-0000-000000000000","name":"TestName","iban":"TEST-IBAN-001"}`,
		},
		{
			name: "unknown JSON field",
			body: fmt.Sprintf(
				`{"merchant_id":%q,"name":"TestName","iban":"TEST-IBAN-001","extra":true}`,
				merchantID,
			),
		},
		{
			name: "multiple JSON objects",
			body: fmt.Sprintf(
				`{"merchant_id":%q,"name":"TestName","iban":"TEST-IBAN-001"} {}`,
				merchantID,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRepo{}
			handler := NewBeneficiaryHandler(repo)

			req := httptest.NewRequest(
				http.MethodPost,
				"/beneficiaries",
				strings.NewReader(tc.body),
			)
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			handler.CreateBeneficiary(recorder, req)

			if recorder.Code != http.StatusBadRequest {
				t.Errorf(
					"status = %d, want %d",
					recorder.Code,
					http.StatusBadRequest,
				)
			}

			if repo.createCalls != 0 {
				t.Errorf("repository calls = %d, want 0", repo.createCalls)
			}
		})
	}
}

func TestBeneficiaryHandler_Create_RepositoryError(t *testing.T) {
	tests := []struct {
		name       string
		repoErr    error
		wantStatus int
		wantBody   string
	}{
		{
			name: "beneficiary already exists",
			repoErr: fmt.Errorf(
				"create beneficiary: %w",
				beneficiary.ErrAlreadyExists,
			),
			wantStatus: http.StatusConflict,
			wantBody:   "beneficiary already exists\n",
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
				createErr: tc.repoErr,
			}
			handler := NewBeneficiaryHandler(repo)

			input := CreateBeneficiaryRequest{
				MerchantID: uuid.New(),
				Name:       "TestName",
				IBAN:       "TEST-IBAN-001",
			}

			body, err := json.Marshal(input)
			if err != nil {
				t.Fatalf("failed to marshal request: %v", err)
			}

			req := httptest.NewRequest(
				http.MethodPost,
				"/beneficiaries",
				bytes.NewReader(body),
			)
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			handler.CreateBeneficiary(recorder, req)

			if recorder.Code != tc.wantStatus {
				t.Errorf(
					"status = %d, want %d",
					recorder.Code,
					tc.wantStatus,
				)
			}

			if got := recorder.Body.String(); got != tc.wantBody {
				t.Errorf(
					"response body = %q, want %q",
					got,
					tc.wantBody,
				)
			}

			if repo.createCalls != 1 {
				t.Errorf(
					"repository calls = %d, want 1",
					repo.createCalls,
				)
			}

			if len(repo.beneficiaries) != 0 {
				t.Errorf(
					"created beneficiaries = %d, want 0",
					len(repo.beneficiaries),
				)
			}
		})
	}
}
