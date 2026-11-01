package postgres

import (
	"context"
	"fmt"

	"github.com/Imanghvs/crudgo/internal/beneficiary"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BeneficiaryRepository struct {
	pool *pgxpool.Pool
}

func NewBeneficiaryRepository(pool *pgxpool.Pool) *BeneficiaryRepository {
	return &BeneficiaryRepository{
		pool: pool,
	}
}

func (r *BeneficiaryRepository) Create(
	ctx context.Context,
	input beneficiary.CreateInput,
) (beneficiary.Beneficiary, error) {
	created := beneficiary.Beneficiary{
		MerchantID: input.MerchantID,
		Name:       input.Name,
		IBAN:       input.IBAN,
	}
	err := r.pool.QueryRow(
		ctx,
		`INSERT INTO beneficiaries (merchant_id, name, iban)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`,
		created.MerchantID,
		created.Name,
		created.IBAN,
	).Scan(&created.ID, &created.CreatedAt)
	if err != nil {
		return beneficiary.Beneficiary{}, fmt.Errorf("create beneficiary: %w", err)
	}

	return created, nil
}
