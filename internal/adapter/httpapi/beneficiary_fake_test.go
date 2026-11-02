package httpapi

import (
	"context"
	"fmt"
	"time"

	"github.com/Imanghvs/crudgo/internal/beneficiary"
	"github.com/google/uuid"
)

type fakeRepo struct {
	beneficiaries []beneficiary.Beneficiary
	createErr     error
	createCalls   int
	getByIDErr    error
	getByIDCalls  int
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

func (r *fakeRepo) GetByID(ctx context.Context, id uuid.UUID) (beneficiary.Beneficiary, error) {
	r.getByIDCalls++

	if r.getByIDErr != nil {
		return beneficiary.Beneficiary{}, r.getByIDErr
	}

	for _, ben := range r.beneficiaries {
		if ben.ID == id {
			return ben, nil
		}
	}

	return beneficiary.Beneficiary{}, fmt.Errorf(
		"get beneficiary by id: %w",
		beneficiary.ErrNotFound,
	)
}
