package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Imanghvs/crudgo/internal/beneficiary"
	"github.com/google/uuid"
)

type BeneficiaryHandler struct {
	repo BeneficiaryRepository
}

type BeneficiaryRepository interface {
	Create(
		ctx context.Context,
		input beneficiary.CreateInput,
	) (beneficiary.Beneficiary, error)
}

func NewBeneficiaryHandler(
	repo BeneficiaryRepository,
) *BeneficiaryHandler {
	return &BeneficiaryHandler{
		repo: repo,
	}
}

type CreateBeneficiaryRequest struct {
	MerchantID uuid.UUID `json:"merchant_id"`
	Name       string    `json:"name"`
	IBAN       string    `json:"iban"`
}

type CreateBeneficiaryResponse struct {
	ID         uuid.UUID `json:"id"`
	MerchantID uuid.UUID `json:"merchant_id"`
	Name       string    `json:"name"`
	IBAN       string    `json:"iban"`
	CreatedAt  time.Time `json:"created_at"`
}

func (handler *BeneficiaryHandler) CreateBeneficiary(
	w http.ResponseWriter,
	r *http.Request,
) {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var dto CreateBeneficiaryRequest
	if err := decoder.Decode(&dto); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		http.Error(w, "request body must contain exactly one JSON object", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(dto.Name) == "" {
		http.Error(w, "name must not be blank", http.StatusBadRequest)
		return
	}

	if dto.MerchantID == uuid.Nil {
		http.Error(w, "invalid or empty merchant id", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(dto.IBAN) == "" {
		http.Error(w, "iban must not be blank", http.StatusBadRequest)
		return
	}

	created, err := handler.repo.Create(
		r.Context(),
		beneficiary.CreateInput{
			MerchantID: dto.MerchantID,
			Name:       dto.Name,
			IBAN:       dto.IBAN,
		},
	)
	if err != nil {
		if errors.Is(err, beneficiary.ErrAlreadyExists) {
			http.Error(w, "beneficiary already exists", http.StatusConflict)
			return
		}

		slog.ErrorContext(r.Context(), "failed to create beneficiary", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	response := CreateBeneficiaryResponse{
		ID:         created.ID,
		MerchantID: created.MerchantID,
		Name:       created.Name,
		IBAN:       created.IBAN,
		CreatedAt:  created.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.ErrorContext(r.Context(), "failed to encode response", "error", err)
	}
}
