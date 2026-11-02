package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"

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
	GetByID(ctx context.Context, id uuid.UUID) (beneficiary.Beneficiary, error)
}

func NewBeneficiaryHandler(
	repo BeneficiaryRepository,
) *BeneficiaryHandler {
	return &BeneficiaryHandler{
		repo: repo,
	}
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

	response := BeneficiaryResponse{
		ID:         created.ID,
		MerchantID: created.MerchantID,
		Name:       created.Name,
		IBAN:       created.IBAN,
		CreatedAt:  created.CreatedAt,
	}

	writeJSON(w, r, http.StatusCreated, response)
}

func (handler *BeneficiaryHandler) GetBeneficiaryByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil || id == uuid.Nil {
		writeHTTPError(w, r, badRequest("invalid beneficiary id"))
		return
	}

	ben, err := handler.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, beneficiary.ErrNotFound) {
			err = &HTTPError{
				Status:  http.StatusNotFound,
				Message: "beneficiary not found",
				Cause:   err,
			}
		}

		writeHTTPError(w, r, err)
		return
	}

	response := BeneficiaryResponse{
		ID:         ben.ID,
		MerchantID: ben.MerchantID,
		Name:       ben.Name,
		IBAN:       ben.IBAN,
		CreatedAt:  ben.CreatedAt,
	}

	writeJSON(w, r, http.StatusOK, response)
}
