package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Imanghvs/crudgo/internal/beneficiary"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "beneficiaries_merchant_iban_unique" {
			return beneficiary.Beneficiary{}, fmt.Errorf(
				"create beneficiary: %w: %w",
				beneficiary.ErrAlreadyExists,
				err,
			)
		}
		return beneficiary.Beneficiary{}, fmt.Errorf("create beneficiary: %w", err)
	}

	return created, nil
}

func (r *BeneficiaryRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (beneficiary.Beneficiary, error) {
	var res beneficiary.Beneficiary
	err := r.pool.QueryRow(
		ctx,
		`SELECT id, merchant_id, name, iban, created_at
		FROM beneficiaries
		WHERE id = $1`,
		id,
	).Scan(&res.ID, &res.MerchantID, &res.Name, &res.IBAN, &res.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return beneficiary.Beneficiary{}, fmt.Errorf(
				"get beneficiary by id: %w: %w",
				beneficiary.ErrNotFound,
				err,
			)
		}
		return beneficiary.Beneficiary{}, fmt.Errorf("get beneficiary by id: %w", err)
	}

	return res, nil
}
