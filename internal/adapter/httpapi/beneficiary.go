package httpapi

import (
	"context"
	"errors"
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
	var dto CreateBeneficiaryRequest

	if err := decodeJSON(w, r, &dto); err != nil {
		writeHTTPError(w, r, err)
		return
	}

	if strings.TrimSpace(dto.Name) == "" {
		writeHTTPError(w, r, badRequest("name must not be blank"))
		return
	}

	if dto.MerchantID == uuid.Nil {
		writeHTTPError(w, r, badRequest("invalid or empty merchant id"))
		return
	}

	if strings.TrimSpace(dto.IBAN) == "" {
		writeHTTPError(w, r, badRequest("iban must not be blank"))
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
			err = &HTTPError{
				Status:  http.StatusConflict,
				Message: "beneficiary already exists",
				Cause:   err,
			}
		}

		writeHTTPError(w, r, err)
		return
	}

	response := CreateBeneficiaryResponse{
		ID:         created.ID,
		MerchantID: created.MerchantID,
		Name:       created.Name,
		IBAN:       created.IBAN,
		CreatedAt:  created.CreatedAt,
	}

	writeJSON(w, r, http.StatusCreated, response)
}
