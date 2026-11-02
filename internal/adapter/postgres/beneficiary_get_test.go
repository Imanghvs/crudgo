package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Imanghvs/crudgo/internal/beneficiary"
	"github.com/google/uuid"
)

func TestBeneficiaryRepository_GetByID(t *testing.T) {
	pool, err := setupDB()
	if err != nil {
		t.Fatalf("test setup failed: %v", err)
	}
	t.Cleanup(pool.Close)

	repo := NewBeneficiaryRepository(pool)

	created := beneficiary.Beneficiary{
		MerchantID: uuid.New(),
		Name:       "TestGetByIdName",
		IBAN:       "TestGetByIdIban",
	}

	insertCtx, cancelInsert := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelInsert()
	err = pool.QueryRow(
		insertCtx,
		`INSERT INTO beneficiaries (merchant_id, name, iban)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`,
		created.MerchantID,
		created.Name,
		created.IBAN,
	).Scan(&created.ID, &created.CreatedAt)
	if err != nil {
		t.Fatalf("failed to insert beneficiary: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		tag, err := pool.Exec(
			ctx,
			`DELETE FROM beneficiaries WHERE id = $1`,
			created.ID,
		)
		if err != nil {
			t.Errorf("failed to clean up beneficiary %s: %v", created.ID, err)
			return
		}

		var expectedRows int64 = 1
		if tag.RowsAffected() != expectedRows {
			t.Errorf(
				"cleanup process expected %d deletions, %d rows deleted",
				expectedRows,
				tag.RowsAffected(),
			)
		}
	})

	getCtx, cancelGet := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelGet()

	retrieved, err := repo.GetByID(
		getCtx,
		created.ID,
	)
	if err != nil {
		t.Fatalf("failed to get beneficiary: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("id = %s, want %s", retrieved.ID, created.ID)
	}
	if !retrieved.CreatedAt.Equal(created.CreatedAt) {
		t.Errorf(
			"created_at = %v, want %v",
			retrieved.CreatedAt,
			created.CreatedAt,
		)
	}
	if retrieved.Name != created.Name {
		t.Errorf("name = %q, want %q", retrieved.Name, created.Name)
	}
	if retrieved.IBAN != created.IBAN {
		t.Errorf("iban = %q, want %q", retrieved.IBAN, created.IBAN)
	}
	if retrieved.MerchantID != created.MerchantID {
		t.Errorf(
			"merchant id = %q, want %q",
			retrieved.MerchantID.String(),
			created.MerchantID.String(),
		)
	}
}

func TestBeneficiaryRepository_GetByID_NotFound(t *testing.T) {
	pool, err := setupDB()
	if err != nil {
		t.Fatalf("test setup failed: %v", err)
	}
	t.Cleanup(pool.Close)

	repo := NewBeneficiaryRepository(pool)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	retrieved, err := repo.GetByID(ctx, uuid.New())

	if !errors.Is(err, beneficiary.ErrNotFound) {
		t.Errorf("error = %v, want ErrNotFound", err)
	}

	if retrieved != (beneficiary.Beneficiary{}) {
		t.Errorf("retrieved = %+v, want zero-value beneficiary", retrieved)
	}
}
