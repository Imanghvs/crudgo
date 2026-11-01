package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Imanghvs/crudgo/internal/beneficiary"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestBeneficiaryRepository_Create(t *testing.T) {
	pool, err := setupDB()
	if err != nil {
		t.Fatalf("test setup failed: %v", err)
	}
	t.Cleanup(pool.Close)

	repo := NewBeneficiaryRepository(pool)

	input := beneficiary.CreateInput{
		Name:       "TestName",
		IBAN:       "TestIBAN",
		MerchantID: uuid.New(),
	}
	createCtx, cancelCreate := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelCreate()
	created, err := repo.Create(
		createCtx,
		input,
	)
	if err != nil {
		t.Fatalf("failed to create beneficiary: %v", err)
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

	if created.ID == uuid.Nil {
		t.Error("beneficiary ID is empty")
	}
	if created.CreatedAt.IsZero() {
		t.Error("created_at is not set")
	}
	if created.Name != input.Name {
		t.Errorf("name = %q, want %q", created.Name, input.Name)
	}
	if created.IBAN != input.IBAN {
		t.Errorf("iban = %q, want %q", created.IBAN, input.IBAN)
	}
	if created.MerchantID != input.MerchantID {
		t.Errorf("merchant id = %q, want %q", created.MerchantID.String(), input.MerchantID.String())
	}

	var stored beneficiary.Beneficiary

	queryCtx, cancelQuery := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancelQuery()
	err = pool.QueryRow(
		queryCtx,
		`SELECT id, merchant_id, name, iban, created_at
		FROM beneficiaries
		WHERE id = $1`,
		created.ID,
	).Scan(
		&stored.ID,
		&stored.MerchantID,
		&stored.Name,
		&stored.IBAN,
		&stored.CreatedAt,
	)

	if err != nil {
		t.Fatalf("failed to fetch created beneficiary from db: %v", err)
	}

	if stored.ID != created.ID ||
		stored.MerchantID != created.MerchantID ||
		stored.Name != created.Name ||
		stored.IBAN != created.IBAN ||
		!stored.CreatedAt.Equal(created.CreatedAt) {
		t.Errorf("persisted beneficiary = %+v, want %+v", stored, created)
	}
}

func setupDB() (*pgxpool.Pool, error) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		return nil, errors.New("TEST_DATABASE_URL not provided")
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to the database: %w", err)
	}

	queryCtx, queryCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer queryCancel()
	var databaseName string
	err = pool.QueryRow(
		queryCtx,
		"SELECT current_database()",
	).Scan(&databaseName)

	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("query current database: %w", err)
	}

	if databaseName != "crudgo_test" {
		pool.Close()
		return nil, fmt.Errorf(
			"unexpected database %q, expected %q",
			databaseName,
			"crudgo_test",
		)
	}

	return pool, nil
}
